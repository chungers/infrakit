package manager

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/leader"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/rpc/client"
	controller_rpc "github.com/docker/infrakit/pkg/rpc/controller"
	group_rpc "github.com/docker/infrakit/pkg/rpc/group"
	rpc "github.com/docker/infrakit/pkg/rpc/group"
	"github.com/docker/infrakit/pkg/run/depends"
	"github.com/docker/infrakit/pkg/spi"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/store"
	"github.com/docker/infrakit/pkg/types"
)

var (
	log = logutil.New("module", "manager")

	debugV  = logutil.V(100)
	debugV2 = logutil.V(500)

	// InterfaceSpec is the current name and version of the Instance API.
	InterfaceSpec = spi.InterfaceSpec{
		Name:    "Manager",
		Version: "0.1.0",
	}
)

// Leadership is the interface for getting information about the current leader node
type Leadership interface {
	// IsLeader returns true only if for certain this is a leader. False if not or unknown.
	IsLeader() (bool, error)
}

// Manager is the interface for interacting locally or remotely with the manager
type Manager interface {
	Leadership

	// LeaderLocation returns the location of the leader
	LeaderLocation() (*url.URL, error)

	// Plan returns the changes to be made
	Plan(specs []types.Spec) (types.Changes, error)

	// Enforce enforces infrastructure state to match that of the specs
	Enforce(specs []types.Spec) error

	// Specs returns the specs this manager tasked with enforcing
	Specs() ([]types.Spec, error)

	// Inspect returns the current state of the infrastructure
	Inspect() ([]types.Object, error)

	// Terminate destroys all resources associated with the specs
	Terminate(specs []types.Spec) error
}

// Backend is the admin / server interface
type Backend interface {
	group.Plugin

	Controllers() (map[string]controller.Controller, error)
	Groups() (map[group.ID]group.Plugin, error)

	Manager

	Start() (<-chan struct{}, error)
	Stop()
}

// manager is the controller of all the plugins.  It is able to process multiple inputs
// such as leadership changes and configuration changes and perform the necessary actions
// to activate / deactivate plugins
type manager struct {
	group.Plugin // Note that some methods are overridden

	plugins     discovery.Plugins
	leader      leader.Detector
	leaderStore leader.Store
	snapshot    store.Snapshot
	isLeader    bool
	lock        sync.Mutex
	stop        chan struct{}
	running     chan struct{}

	backendName string
	backendOps  chan<- backendOp
}

type backendOp struct {
	name      string
	operation func() error
}

// NewManager returns the manager which depends on other services to coordinate and manage
// the plugins in order to ensure the infrastructure state matches the user's spec.
func NewManager(plugins discovery.Plugins,
	leader leader.Detector,
	leaderStore leader.Store,
	snapshot store.Snapshot,
	backendName string) Backend {

	return &manager{
		// "base class" is the stateless backend group plugin
		Plugin: &lateBindGroup{
			finder: func() (group.Plugin, error) {
				endpoint, err := plugins.Find(plugin.Name(backendName))
				if err != nil {
					return nil, err
				}
				return rpc.NewClient(endpoint.Address)
			},
		},
		plugins:     plugins,
		leader:      leader,
		leaderStore: leaderStore,
		snapshot:    snapshot,
		backendName: backendName,
	}
}

// return true only if the current call caused an allocation of the running channel.
func (m *manager) initRunning() bool {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.running == nil {
		m.running = make(chan struct{})
		return true
	}
	return false
}

// IsLeader returns leader status.  False if not or unknown.
func (m *manager) IsLeader() (bool, error) {
	return m.isLeader, nil
}

// LeaderLocation returns the location of the leader
func (m *manager) LeaderLocation() (*url.URL, error) {
	if m.leaderStore == nil {
		return nil, fmt.Errorf("cannot locate leader")
	}

	return m.leaderStore.GetLocation()
}

// Plan returns the changes needed given the new input
func (m *manager) Plan(specs []types.Spec) (types.Changes, error) {
	current, err := m.Specs()
	if err != nil {
		return types.Changes{}, err
	}

	currentSpecs := types.Specs(current)
	updatedSpecs := types.Specs(specs)
	return currentSpecs.Changes(updatedSpecs), nil
}

// Enforce enforces infrastructure state to match that of the specs
func (m *manager) Enforce(specs []types.Spec) error {

	// TODO
	requested := globalSpec{}
	for _, s := range specs {
		handler := plugin.NameFrom(s.Kind, s.Metadata.Name)
		if s.Kind == "group" {
			// TODO(chungers) -- this really needs to be cleaned up
			handler = plugin.Name(m.backendName)
			gspec := group.Spec{
				ID:         group.ID(s.Metadata.Name),
				Properties: s.Properties,
			}
			requested.updateGroupSpec(gspec, handler)
		} else {
			requested.updateSpec(s, handler)
		}
	}

	// Note we also have a version that's in the persistent store.
	// Should we do some delta calculations?

	return requested.store(m.snapshot)
}

// Specs returns the specs this manager is tasked with enforcing
func (m *manager) Specs() ([]types.Spec, error) {
	global := globalSpec{}
	err := global.load(m.snapshot)
	if err != nil {
		return nil, err
	}
	saved := []types.Spec{}
	err = global.visit(func(k key, r record) error {
		saved = append(saved, r.Spec)
		return nil
	})
	return saved, err
}

func (m *manager) allControllers(work func(controller.Controller) error) error {
	// Go through all the controllers
	running, err := m.plugins.List()
	if err != nil {
		return err
	}
	log.Debug("current state", "running", running, "manager", m, "V", debugV2)
	for lookup, endpoint := range running {
		rpcClient, err := client.New(endpoint.Address, controller.InterfaceSpec)
		log.Debug("Scanned controller", "name", lookup, "at", endpoint, "err", err, "manager", m, "V", debugV2)
		if err == nil {
			name := plugin.Name(lookup)
			log.Debug("Calling controller", "name", name, "V", debugV2)
			c := controller_rpc.Adapt(name, rpcClient)
			err = work(c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Inspect returns the current state of the infrastructure.  It performs an 'all-shard' query across
// all plugins of the type 'group' and then aggregate the results.
func (m *manager) Inspect() ([]types.Object, error) {
	aggregated := []types.Object{}
	err := m.allControllers(func(c controller.Controller) error {
		objects, err := c.Describe(nil)
		if err != nil {
			return err
		}
		aggregated = append(aggregated, objects...)
		return nil
	})
	return aggregated, err
}

// Terminate destroys all resources associated with the specs
func (m *manager) Terminate(specs []types.Spec) error {
	return fmt.Errorf("not implemented")
}

// Start starts the manager.  It does not block. Instead read from the returned channel to block.
func (m *manager) Start() (<-chan struct{}, error) {

	// initRunning guarantees that the m.running will be initialized the first time it's
	// called.  If another call of Start is made after the first, don't do anything just return the references.
	if !m.initRunning() {
		return m.running, nil
	}

	log.Info("Manager starting")

	leaderChan, err := m.leader.Start()
	if err != nil {
		return nil, err
	}

	m.stop = make(chan struct{})
	notify := make(chan bool)
	stopWorkQueue := make(chan struct{})

	// proxied backend needs to have its operations serialized with respect to leadership calls, etc.
	backendOps := make(chan backendOp)
	m.backendOps = backendOps

	// This goroutine here serializes work so that we don't have concurrent commits or unwatches / updates / etc.
	go func() {

		for {
			select {

			case op := <-backendOps:
				log.Debug("Backend operation", "op", op, "V", debugV2)
				if m.isLeader {
					op.operation()
				}

			case <-stopWorkQueue:

				log.Info("Stopping work queue.")
				close(m.running)
				log.Info("Manager stopped.")
				return

			case leader, open := <-notify:

				if !open {
					return
				}

				// This channel has data only when there's been a leadership change.

				log.Debug("leader event", "leader", leader)
				if leader {
					m.onAssumeLeadership()
				} else {
					m.onLostLeadership()
				}
			}

		}
	}()

	// Goroutine for handling all inbound leadership and control events.
	go func() {

		for {
			select {

			case <-m.stop:

				log.Info("Stopping..")
				m.stop = nil
				close(stopWorkQueue)
				close(notify)
				return

			case evt, open := <-leaderChan:

				if !open {
					return
				}

				// This here handles possible duplicated events about leadership and fires only when there
				// is a change.

				m.lock.Lock()

				current := m.isLeader

				if evt.Status == leader.Unknown {
					log.Warn("Leadership status is uncertain", "err", evt.Error)

					// if we are currently the leader then there's a possibility of split brain depending on
					// the robustness of the leader election process.
					// It's better to be conservative and not assume we can still be a leader...  just downgrade
					// because the worst case is we stopped watching (and not have two masters running wild).

					if m.isLeader {
						m.isLeader = false
					}

				} else {
					m.isLeader = evt.Status == leader.Leader
				}
				next := m.isLeader

				m.lock.Unlock()

				if current != next {
					notify <- next
				}

			}
		}

	}()

	return m.running, nil
}

// Stop stops the manager
func (m *manager) Stop() {
	if m.stop == nil {
		return
	}
	close(m.stop)
	m.leader.Stop()
}

// func (m *manager) getCurrentState() (globalSpec, error) {
// 	// TODO(chungers) -- using the group plugin backend here isn't the general case.
// 	// When plugin activation is implemented, it's possible to have multiple group plugins
// 	// and the only way to reconstruct the globalSpec, which contains multiple groups of
// 	// possibly different group plugin implementations, is to do an 'all-shard' query across
// 	// all plugins of the type 'group' and then aggregate the results into the final globalSpec.
// 	// For now this just uses the gross simplification of asking the group plugin that the manager
// 	// proxies.

// 	global := globalSpec{}
// 	// Go through all the controllers
// 	running, err := m.plugins.List()
// 	if err != nil {
// 		return global, err
// 	}
// 	log.Debug("current state", "running", running, "manager", m)
// 	for lookup, endpoint := range running {
// 		rpcClient, err := client.New(endpoint.Address, controller.InterfaceSpec)
// 		if err == nil {
// 			c := controller_rpc.Adapt(plugin.Name(lookup), rpcClient)
// 			objects, err := c.Describe(nil)
// 			log.Debug("Describe on controller", "controller", c, "name", lookup, "objects", objects, "err", err, "manager", m)
// 			if err != nil {
// 				log.Error("Error describing controller", "c", c, "err", err, "manager", m)
// 				return global, err
// 			}
// 			for _, object := range objects {
// 				global.updateSpec(object.Spec, plugin.NameFrom(lookup, object.Spec.Metadata.Name))
// 			}
// 		}
// 	}

// 	return global, nil
// }

func (m *manager) checkPluginsRunning(config *globalSpec) (<-chan struct{}, error) {
	wait := make(chan struct{})

	specs := config.specs()
	runnables, err := depends.RunnablesFrom(specs)
	if err != nil {
		return nil, err
	}

	if len(runnables) == 0 {
		// There's nothing to do. Don't wait just return
		close(wait)
		return wait, nil
	}

	go func(check depends.Runnables, plugins discovery.Plugins) {

		for {
			running, err := plugins.List()
			if err != nil {
				log.Error("Cannot list running plugins", "err", err)
				continue
			}

			match := 0
			for _, runnable := range runnables {
				lookup, _ := runnable.Plugin().GetLookupAndType()
				if _, has := running[lookup]; has {
					match++
				}
			}
			if match == len(runnables) {
				close(wait)
				return
			}

			<-time.After(5 * time.Second)
		}

	}(runnables, m.plugins)
	return wait, nil
}

func (m *manager) onAssumeLeadership() error {
	log.Info("Assuming leadership")

	// load the config
	config := &globalSpec{}
	// load the latest version -- assumption here is that it's been persisted already.
	log.Info("Loading snapshot")
	err := config.load(m.snapshot)
	log.Info("Loaded snapshot", "err", err)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return err
	}

	// check that all the plugins referenced in the config are running, block here
	wait, err := m.checkPluginsRunning(config)
	if err != nil {
		return err
	}

	<-wait

	log.Info("Committing specs to controllers", "config", config)

	return m.callControllers(*config,
		func(c controller.Controller, spec types.Spec) error {
			log.Info("Committing spec", "controller", c, "spec", spec)
			_, err := c.Commit(controller.Enforce, spec)
			if err != nil {
				log.Error("Error committing spec", "spec", spec, "err", err)
			}
			return err
		},
		func(c group.Plugin, gspec group.Spec) error {
			log.Info("Committing group spec", "groupPlugin", c, "spec", gspec)
			_, err := c.CommitGroup(gspec, false)
			if err != nil {
				log.Error("Error committing group spec", "spec", gspec, "err", err)
			}
			return err
		})
}

func (m *manager) onLostLeadership() error {
	log.Info("Lost leadership")

	// load the config
	config := &globalSpec{}
	// load the latest version -- assumption here is that it's been persisted already.
	log.Info("Loading snapshot")
	err := config.load(m.snapshot)
	log.Info("Loaded snapshot", "err", err)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return err
	}

	// check that all the plugins referenced in the config are running, block here
	wait, err := m.checkPluginsRunning(config)
	if err != nil {
		return err
	}

	<-wait

	return m.allControllers(func(c controller.Controller) error {
		log.Debug("Freeing controller", "c", c)
		_, err := c.Free(nil)
		return err
	})
}

func (m *manager) callControllers(config globalSpec,
	callController func(controller.Controller, types.Spec) error,
	callGroupPlugin func(group.Plugin, group.Spec) error,
) error {

	running, err := m.plugins.List()
	if err != nil {
		return err
	}

	log.Debug("running", "running", running, "index", config.index)

	return config.visit(func(k key, r record) error {
		// for controllers, the kind and the name make up to be the plugin name
		pn := plugin.Name(k.Name)
		if !r.Handler.IsEmpty() {
			pn = r.Handler
		}
		interfaceSpec := r.InterfaceSpec
		if r.Spec.Version != "" {
			interfaceSpec = spi.DecodeInterfaceSpec(r.Spec.Version)
		}
		lookup, _ := pn.GetLookupAndType()
		endpoint, has := running[lookup]
		if !has {
			return nil
		}
		log.Debug("Dispatching work", "name", pn, "interfaceSpec", interfaceSpec, "lookup", lookup)
		rpcClient, err := client.New(endpoint.Address, interfaceSpec)
		if err == nil {
			switch interfaceSpec {
			case controller.InterfaceSpec:
				c := controller_rpc.Adapt(plugin.Name(lookup), rpcClient)
				log.Debug("Calling on controller", "controller", c, "name", pn, "spec", r.Spec)
				err = callController(c, r.Spec)
				if err != nil {
					log.Error("Error running on controller", "c", c, "spec", r.Spec, "err", err)
					return err
				}

			case group.InterfaceSpec:
				c := group_rpc.Adapt(rpcClient)
				log.Debug("Calling on group", "controller", c, "name", pn, "spec", r.Spec)
				err = callGroupPlugin(c, group.Spec{ID: group.ID(r.Spec.Metadata.Name), Properties: r.Spec.Properties})
				if err != nil {
					log.Error("Error running on group", "c", c, "spec", r.Spec, "err", err)
					return err
				}
			}
		}
		return nil
	})
}

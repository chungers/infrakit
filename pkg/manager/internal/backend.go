package internal

import (
	"fmt"
	"sync"
	"time"

	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/core"
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
)

// Backend is the controller of all the plugins.  It is able to process multiple inputs
// such as leadership changes and configuration changes and perform the necessary actions
// to activate / deactivate plugins
type Backend struct {
	self         plugin.Name // the name this backend is listening on.
	group.Plugin             // TODO - remove this

	plugins     discovery.Plugins
	leader      leader.Detector
	leaderStore leader.Store
	snapshot    store.Snapshot
	isLeader    bool
	lock        sync.Mutex
	stop        chan struct{}
	running     chan struct{}

	backendName string
	//	backendOps  <-chan backendOp
	fanin *fanin
}

type backendOp struct {
	context   interface{}
	name      string
	operation func() error
}

// NewBackend returns the manager which depends on other services to coordinate and manage
// the plugins in order to ensure the infrastructure state matches the user's spec.
func NewBackend(name plugin.Name,
	plugins discovery.Plugins,
	leader leader.Detector,
	leaderStore leader.Store,
	snapshot store.Snapshot,
	backendName string) *Backend {

	return &Backend{
		self: name,
		// "base class" is the stateless backend group plugin
		Plugin: group.LazyConnect(
			func() (group.Plugin, error) {
				endpoint, err := plugins.Find(plugin.Name(backendName))
				if err != nil {
					return nil, err
				}
				return rpc.NewClient(endpoint.Address)
			}, 5*time.Second),
		plugins:     plugins,
		leader:      leader,
		leaderStore: leaderStore,
		snapshot:    snapshot,
		backendName: backendName,
	}
}

// return true only if the current call caused an allocation of the running channel.
func (b *Backend) initRunning() bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	if b.running == nil {
		b.running = make(chan struct{})
		return true
	}
	return false
}

type visitor struct {
	interfaceSpec spi.InterfaceSpec
	work          func(plugin.Name, client.Client) error
}

func asController(v func(plugin.Name, controller.Controller) error) visitor {
	return visitor{
		interfaceSpec: controller.InterfaceSpec,
		work: func(n plugin.Name, rpc client.Client) error {
			return v(n, controller_rpc.Adapt(n, rpc))
		},
	}
}

func asGroupPlugin(v func(plugin.Name, group.Plugin) error) visitor {
	return visitor{
		interfaceSpec: group.InterfaceSpec,
		work: func(n plugin.Name, rpc client.Client) error {
			return v(n, group_rpc.Adapt(rpc))
		},
	}
}

func (b *Backend) visitPlugins(visitors ...visitor) error {
	// Go through all the controllers
	running, err := b.plugins.List()
	if err != nil {
		return err
	}

	self, _ := b.self.GetLookupAndType()

	log.Debug("current state", "running", running, "backend", b, "V", debugV2)
	for lookup, endpoint := range running {

		if lookup == self {
			continue
		}

		for _, visitor := range visitors {
			rpcClient, err := client.New(endpoint.Address, visitor.interfaceSpec)
			log.Debug("Scanned controller", "name", lookup, "at", endpoint, "err", err, "backend", b, "V", debugV2)
			if err == nil {
				name := plugin.Name(lookup)
				log.Debug("Calling controller", "name", name, "V", debugV2)
				err = visitor.work(name, rpcClient)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Start starts the manager.  It does not block. Instead read from the returned channel to block.
func (b *Backend) Start() (<-chan struct{}, error) {

	// initRunning guarantees that the b.running will be initialized the first time it's
	// called.  If another call of Start is made after the first, don't do anything just return the references.
	if !b.initRunning() {
		return b.running, nil
	}

	log.Info("Manager starting")

	leaderChan, err := b.leader.Start()
	if err != nil {
		return nil, err
	}

	b.stop = make(chan struct{})
	notify := make(chan bool)
	stopWorkQueue := make(chan struct{})

	// proxied backend needs to have its operations serialized with respect to leadership calls, etc.
	done := make(chan struct{})
	fanin, backendOps := newFanin(done)

	//b.backendOps = backendOps
	b.fanin = fanin

	// This goroutine here serializes work so that we don't have concurrent commits or unwatches / updates / etc.
	go func() {

		for {
			select {

			case op := <-backendOps:
				log.Debug("Backend operation", "op", op, "V", debugV2)
				if b.isLeader {
					op.operation()
				}

			case <-stopWorkQueue:

				log.Info("Stopping work queue.")
				close(done)
				close(b.running)
				log.Info("Manager stopped.")
				return

			case leader, open := <-notify:

				if !open {
					return
				}

				// This channel has data only when there's been a leadership change.

				log.Debug("leader event", "leader", leader)
				if leader {
					b.onAssumeLeadership()
				} else {
					b.onLostLeadership()
				}
			}

		}
	}()

	// Goroutine for handling all inbound leadership and control events.
	go func() {

		for {
			select {

			case <-b.stop:

				log.Info("Stopping..")
				b.stop = nil
				close(stopWorkQueue)
				close(notify)
				return

			case evt, open := <-leaderChan:

				if !open {
					return
				}

				// This here handles possible duplicated events about leadership and fires only when there
				// is a change.

				b.lock.Lock()

				current := b.isLeader

				if evt.Status == leader.Unknown {
					log.Warn("Leadership status is uncertain", "err", evt.Error)

					// if we are currently the leader then there's a possibility of split brain depending on
					// the robustness of the leader election process.
					// It's better to be conservative and not assume we can still be a leader...  just downgrade
					// because the worst case is we stopped watching (and not have two masters running wild).

					if b.isLeader {
						b.isLeader = false
					}

				} else {
					b.isLeader = evt.Status == leader.Leader
				}
				next := b.isLeader

				b.lock.Unlock()

				if current != next {
					notify <- next
				}

			}
		}

	}()

	return b.running, nil
}

// Stop stops the manager
func (b *Backend) Stop() {
	if b.stop == nil {
		return
	}
	close(b.stop)
	b.leader.Stop()
}

func (b *Backend) checkPluginsRunning(config *globalSpec) (<-chan struct{}, error) {
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

	}(runnables, b.plugins)
	return wait, nil
}

func (b *Backend) onAssumeLeadership() error {
	log.Info("Assuming leadership")

	// load the config
	config := &globalSpec{}
	// load the latest version -- assumption here is that it's been persisted already.
	log.Info("Loading snapshot")
	err := config.load(b.snapshot)
	log.Info("Loaded snapshot", "err", err)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return err
	}

	// check that all the plugins referenced in the config are running, block here
	wait, err := b.checkPluginsRunning(config)
	if err != nil {
		return err
	}

	<-wait

	log.Info("Committing specs to controllers", "config", config)

	return b.visitPlugins(asController(
		func(n plugin.Name, c controller.Controller) error {

			spec, has := config.findSpec(n)
			if !has {
				log.Debug("no spec found", "plugin", n, "V", debugV2)
				return nil
			}
			log.Info("Committing spec", "controller", c, "spec", spec)
			_, err := c.Commit(controller.Enforce, spec)
			if err != nil {
				log.Error("Error committing spec", "spec", spec, "err", err)
			}
			return err
		},
	))
}

func (b *Backend) onLostLeadership() error {
	log.Info("Lost leadership")

	// load the config
	config := &globalSpec{}
	// load the latest version -- assumption here is that it's been persisted already.
	log.Info("Loading snapshot")
	err := config.load(b.snapshot)
	log.Info("Loaded snapshot", "err", err)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return err
	}

	// check that all the plugins referenced in the config are running, block here
	wait, err := b.checkPluginsRunning(config)
	if err != nil {
		return err
	}

	<-wait

	return b.visitPlugins(asController(
		func(n plugin.Name, c controller.Controller) error {
			log.Debug("Pausing controller", "c", c)
			_, err := c.Pause(nil)
			return err
		},
	))
}

func (b *Backend) callControllers(config globalSpec,
	callController func(controller.Controller, types.Spec) error,
	callGroupPlugin func(group.Plugin, group.Spec) error,
) error {

	running, err := b.plugins.List()
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

// AllGroupSpecs returns all the group specs stored
func (b *Backend) AllGroupSpecs() ([]group.Spec, error) {
	// load the config
	config := globalSpec{}

	// load the latest version -- assumption here is that it's been persisted already.
	err := config.load(b.snapshot)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return nil, err
	}
	return config.allGroupSpecs()
}

// FindGroupSpec returns the group spec given ID
func (b *Backend) FindGroupSpec(id group.ID) (group.Spec, error) {
	// load the config
	config := globalSpec{}

	// load the latest version -- assumption here is that it's been persisted already.
	err := config.load(b.snapshot)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return group.Spec{}, err
	}
	return config.getGroupSpec(id)
}

// UpdateGroupSpec updates the given spec in storage
func (b *Backend) UpdateGroupSpec(spec group.Spec) error {
	log.Debug("Updating config", "spec", spec)
	b.lock.Lock()
	defer b.lock.Unlock()

	// Always read and then update with the current value.  Assumes the user's input
	// is always authoritative.
	stored := globalSpec{}
	err := stored.store(b.snapshot)
	if err != nil {
		return err
	}

	stored.updateGroupSpec(spec, plugin.Name(b.backendName))
	log.Debug("Saving updated config", "config", stored)

	return stored.store(b.snapshot)
}

// RemoveGroupSpec removes the spec by group id
func (b *Backend) RemoveGroupSpec(id group.ID) error {
	log.Debug("Removing config", "groupID", id)
	b.lock.Lock()
	defer b.lock.Unlock()

	// Always read and then update with the current value.  Assumes the user's input
	// is always authoritative.
	stored := globalSpec{}
	err := stored.load(b.snapshot)
	if err != nil {
		return err
	}

	stored.removeGroup(id)
	log.Debug("Saving updated config", "config", stored)

	return stored.store(b.snapshot)
}

// FindSpecs returns matching specs given the search (can be nil to select all).
func (b *Backend) FindSpecs(search *types.Metadata) (found []types.Spec, err error) {
	// load the config
	config := globalSpec{}
	err = config.load(b.snapshot)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return
	}
	match := func(other types.Metadata) bool {
		if search == nil {
			return true
		}
		copy := *search
		return copy.Compare(other) == 0
	}

	found = []types.Spec{}
	for _, spec := range config.specs() {
		if match(spec.Metadata) {
			found = append(found, spec)
		}
	}
	return
}

// UpdateSpec updates the spec from storage
func (b *Backend) UpdateSpec(spec types.Spec) error {
	config := globalSpec{}
	err := config.load(b.snapshot)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return err
	}
	addressable := core.AsAddressable(&spec)
	config.updateSpec(spec, addressable.Plugin())
	return config.store(b.snapshot)
}

// RemoveSpec removes the spec from storage
func (b *Backend) RemoveSpec(spec types.Spec) error {
	config := globalSpec{}
	err := config.load(b.snapshot)
	if err != nil {
		log.Warn("Error loading config", "err", err)
		return err
	}
	config.removeSpec(spec.Kind, spec.Metadata)
	return config.store(b.snapshot)
}

func (b *Backend) Groups() (map[group.ID]group.Plugin, error) {
	groups := map[group.ID]group.Plugin{}
	err := b.visitPlugins(
		asGroupPlugin(
			func(n plugin.Name, g group.Plugin) error {
				all, err := g.InspectGroups()
				if err != nil {
					return err
				}
				groups[group.ID(n.Lookup())] = g
				for _, gspec := range all {
					gid := group.ID(n.WithType(gspec.ID).String())
					groups[gid] = g
				}
				return nil
			},
		),
	)
	return groups, err
}

// Controllers returns a map of *scoped* controllers across all the discovered
// controllers.
func (b *Backend) Controllers() (map[string]controller.Controller, error) {
	// Addressability -- we need to "shift 1" from this namespace down to indivdually addressable
	// objects in the destination controller
	controllers := map[string]controller.Controller{}
	err := b.visitPlugins(
		asController(
			func(n plugin.Name, c controller.Controller) error {
				all, err := c.Specs(nil)
				if err != nil {
					return err
				}
				controllers[n.Lookup()] = c
				for _, spec := range all {

					addr := core.AddressableFromSpec(spec)
					k := addr.Plugin()
					if !k.HasType() {
						k = n.WithType(k.Lookup())
					}
					fmt.Println(">>>> BACKEND", k)
					controllers[k.String()] = c
				}
				return nil
			},
		),
	)
	return controllers, err
}

// Groups returns a map of *scoped* group controllers by ID of the group.
func (b *Backend) Groups0() (map[group.ID]group.Plugin, error) {
	groups := map[group.ID]group.Plugin{
		group.ID(""): b,
	}
	all, err := b.Plugin.InspectGroups()
	if err != nil {
		return groups, nil
	}
	for _, spec := range all {
		gid := spec.ID
		groups[gid] = b
	}
	log.Debug("Groups", "map", groups, "V", debugV2)
	return groups, nil
}

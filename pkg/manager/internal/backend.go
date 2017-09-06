package internal

import (
	"fmt"
	"path"
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
	//	rpc "github.com/docker/infrakit/pkg/rpc/group"
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
	self plugin.Name // the name this backend is listening on.

	plugins     discovery.Plugins
	leader      leader.Detector
	leaderStore leader.Store
	snapshot    store.Snapshot
	isLeader    bool
	lock        sync.Mutex
	stop        chan struct{}
	running     chan struct{}

	rpcClients     map[rpcClientKey]interface{}
	rpcClientsLock sync.RWMutex

	fanin *fanin
}

type rpcClientKey struct {
	lookup        string
	interfaceSpec spi.InterfaceSpec
}

func (b *Backend) getClient(lookup string, interfaceSpec spi.InterfaceSpec) interface{} {
	b.rpcClientsLock.RLock()
	defer b.rpcClientsLock.RUnlock()
	key := rpcClientKey{lookup: lookup, interfaceSpec: interfaceSpec}
	return b.rpcClients[key]
}

func (b *Backend) setClient(lookup string, interfaceSpec spi.InterfaceSpec, adapter interface{}) {
	b.rpcClientsLock.Lock()
	defer b.rpcClientsLock.Unlock()
	key := rpcClientKey{lookup: lookup, interfaceSpec: interfaceSpec}
	b.rpcClients[key] = adapter
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
	snapshot store.Snapshot) *Backend {

	return &Backend{
		self:        name,
		plugins:     plugins,
		leader:      leader,
		leaderStore: leaderStore,
		snapshot:    snapshot,
		rpcClients:  map[rpcClientKey]interface{}{},
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
	work          func(plugin.Name, interface{}) error
	adapt         func(*Backend, plugin.Name, client.Client) interface{}
}

func asController(v func(plugin.Name, controller.Controller) error) visitor {
	return visitor{
		interfaceSpec: controller.InterfaceSpec,
		work: func(n plugin.Name, rpc interface{}) error {
			return v(n, rpc.(controller.Controller))
		},
		adapt: func(b *Backend, n plugin.Name, rpc client.Client) interface{} {
			c := controller_rpc.Adapt(n, rpc)

			// use internal queued controller
			ch := make(chan backendOp)

			queuedController := QueuedController(c, ch,
				b.FindSpecs,
				b.UpdateSpec,
				b.RemoveSpec)

			b.fanin.add(ch)

			return queuedController
		},
	}
}

func asGroupPlugin(v func(plugin.Name, group.Plugin) error) visitor {
	return visitor{
		interfaceSpec: group.InterfaceSpec,
		work: func(n plugin.Name, rpc interface{}) error {
			return v(n, rpc.(group.Plugin))
		},
		adapt: func(b *Backend, n plugin.Name, rpc client.Client) interface{} {

			g := group_rpc.Adapt(rpc)

			// use internal queued plugin
			ch := make(chan backendOp)

			queuedGroupPlugin := QueuedGroupPlugin(n, g, ch,
				b.AllGroupSpecs,
				b.FindGroupSpec,
				b.UpdateGroupSpec,
				b.RemoveGroupSpec)

			b.fanin.add(ch)

			return queuedGroupPlugin
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

	visit:
		for _, visitor := range visitors {

			name := plugin.Name(lookup)

			// check if we have a stored rpcClient
			adapter := b.getClient(lookup, visitor.interfaceSpec)
			if adapter == nil {
				rpcClient, err := client.New(endpoint.Address, visitor.interfaceSpec)
				if err != nil {
					// not the right interface
					continue visit
				}

				adapter = visitor.adapt(b, name, rpcClient)
				b.setClient(lookup, visitor.interfaceSpec, adapter)
				log.Debug("Cached adapter", "name", name, "interfaceSpec", visitor.interfaceSpec, "V", debugV2)

			}

			// Now we have a matched client
			log.Debug("Found adapter", "name", name, "adapater", adapter, "V", debugV2)
			if err := visitor.work(name, adapter); err != nil {
				return err
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
	if true {
		// REMOVE ME
		return nil
	}

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

	stored.updateGroupSpec(spec, b.self)
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
	addressable := core.AsAddressable(spec)
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

// Groups returns a map of group plugins managed by this backend.
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

					if false {
						//gid := path.Base(string(gspec.ID)) //string(gspec.ID)
						gid := string(gspec.ID)
						fmt.Println(">>>>", gid, "base", path.Base(string(gspec.ID)))

						// need to base the gspec.ID since it may have /
						id := group.ID(n.WithType(gid).String())

						groups[id] = g
					}

					groups[group.ID(n.LookupOnly().WithType(path.Base(string(gspec.ID))))] = g
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

					addr := core.AsAddressable(spec)
					k := addr.Plugin()
					if !k.HasType() {
						k = n.WithType(k.Lookup())
					}
					controllers[k.String()] = c
				}
				return nil
			},
		),
	)
	return controllers, err
}

package manager

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/discovery"
	"github.com/docker/infrakit/launch"
	"github.com/docker/infrakit/leader"
	rpc "github.com/docker/infrakit/rpc/group"
	"github.com/docker/infrakit/spi/group"
	"github.com/docker/infrakit/store"
)

// Service is the service interface for the manager
type Service interface {

	// Also implements the single group plugin interface
	group.Plugin

	// Commit is the operation for multiple groups
	Commit() error
}

// Backend is the
type Backend interface {
	Service

	Start() (<-chan struct{}, error)
	Stop()
}

// manager is the controller of all the plugins.  It is able to process multiple inputs
// such as leadership changes and configuration changes and perform the necessary actions
// to activate / deactivate plugins
type manager struct {
	group.Plugin

	launcher      launch.Launcher
	plugins       discovery.Plugins
	leader        leader.Detector
	snapshot      store.Snapshot
	isLeader      bool
	lock          sync.Mutex
	stop          chan struct{}
	running       chan struct{}
	commit        chan struct{}
	currentConfig GlobalSpec

	backendName string
}

// NewManager returns the manager which depends on other services to coordinate and manage
// the plugins in order to ensure the infrastructure state matches the user's spec.
func NewManager(launcher launch.Launcher, plugins discovery.Plugins, leader leader.Detector, snapshot store.Snapshot,
	backendName string) (Backend, error) {

	m := &manager{
		launcher: launcher,
		plugins:  plugins,
		leader:   leader,
		snapshot: snapshot,
	}

	gp, err := m.proxyForGroupPlugin(backendName)
	if err != nil {
		return nil, err
	}

	m.Plugin = gp
	return m, nil
}

// Start starts the manager.  It does not block. Instead read from the returned channel to block.
func (m *manager) Start() (<-chan struct{}, error) {
	m.lock.Lock()
	if m.running != nil {
		m.lock.Unlock()
		return m.running, nil
	}

	leaderChan, err := m.leader.Start()
	if err != nil {
		m.lock.Unlock()
		return nil, err
	}

	m.running = make(chan struct{})
	m.stop = make(chan struct{})
	m.commit = make(chan struct{})

	// don't hold the lock forever.
	m.lock.Unlock()

	notify := make(chan bool)
	stopWorkQueue := make(chan struct{})

	go func() {
		// This goroutine here serializes work so that we don't have concurrent commits or unwatches
		for {
			select {

			case <-stopWorkQueue:

				log.Infoln("Stopping work queue.")
				close(m.running)
				log.Infoln("Manager stopped.")
				return

			case leader := <-notify:

				log.Debugln("leader:", leader)
				if leader {
					m.onAssumeLeadership()
				} else {
					m.onLostLeadership()
				}

			case <-m.commit:

				if m.isLeader {

					log.Infoln("Begin commit")
					m.doCommit()
					log.Infoln("End commit")

				} else {
					log.Warningln("Try to commit but not leader. No action")
				}
			}
		}
	}()

	go func() {

		for {
			select {

			case <-m.stop:

				log.Infoln("Stopping..")
				close(notify)
				close(stopWorkQueue)
				return

			case evt := <-leaderChan:

				m.lock.Lock()

				current := m.isLeader

				if evt.Status == leader.StatusUnknown {
					log.Warningln("Leadership status is uncertain:", evt.Error)

					// if we are currently the leader then there's a possibility of split brain depending on
					// the robustness of the leader election process.
					// It's better to be conservative and not assume we can still be a leader...  just downgrade
					// because the worst case is we stopped watching (and not have two masters running wild).

					if m.isLeader {
						m.isLeader = false
					}

				} else {
					m.isLeader = evt.Status == leader.StatusLeader
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
	m.leader.Stop()
	close(m.stop)
}

// Commit triggers a snapshot of the config tree to be taken and with that, the manager will
// begin activating and updating the plugins, if it is running as the leader.
func (m *manager) Commit() error {
	if m.commit != nil {
		m.commit <- struct{}{}
	}
	return nil
}

func (m *manager) onAssumeLeadership() error {
	log.Infoln("Assuming leadership")

	// load the config
	config := GlobalSpec{}

	// load the latest version -- assumption here is that it's been persisted already.
	err := m.snapshot.Load(&config)
	if err != nil {
		log.Warningln("Error loading config", err)
		return err
	}

	log.Infoln("Loaded snapshot. err=", err)
	if err != nil {
		return err
	}

	m.lock.Lock()
	m.currentConfig = config // make a copy
	m.lock.Unlock()

	return m.doWatchGroups(config)
}

func (m *manager) onLostLeadership() error {
	log.Infoln("Lost leadership")

	// Unwatch uses a cached version in memory.  This means that we are never able to
	// unwatch without first watching.  If the manager crashes and comes back up again,
	// there will be no possibility to unwatch existing groups.
	m.lock.Lock()
	config := m.currentConfig
	m.lock.Unlock()

	return m.doUnwatchGroups(config)
}

func (m *manager) doCommit() error {

	// load the config
	config := GlobalSpec{}

	// load the latest version -- assumption here is that it's been persisted already.
	err := m.snapshot.Load(&config)
	if err != nil {
		return err
	}

	log.Infoln("Committing.  Loaded snapshot. err=", err)
	if err != nil {
		return err
	}

	m.lock.Lock()
	m.currentConfig = config // make a copy
	m.lock.Unlock()

	return m.doUpdateGroups(config)
}

func (m *manager) doUpdateGroups(config GlobalSpec) error {
	err := m.launchPlugins(config)
	if err != nil {
		return err
	}

	return m.execPlugins(config,
		func(plugin group.Plugin, spec group.Spec) error {

			log.Infoln("UPDATE group", spec.ID, "with spec:", spec)
			err := plugin.UpdateGroup(spec)

			// TODO(chungers) -- yes this is clunky
			if err != nil && strings.Contains(err.Error(), "not being watched") {

				log.Infoln("UPDATE group", spec.ID, "changed to WATCH")
				err = plugin.WatchGroup(spec)

			}

			if err != nil {
				log.Warningln("Error updating/watch group:", spec.ID, "Err=", err)
			}
			return err
		})
}

func (m *manager) doWatchGroups(config GlobalSpec) error {
	err := m.launchPlugins(config)
	if err != nil {
		log.Warningln("Error starting up plugins", err)
		return err
	}
	log.Infoln("Start watching groups")
	return m.execPlugins(config,
		func(plugin group.Plugin, spec group.Spec) error {

			log.Infoln("WATCH group", spec.ID, "with spec:", spec)
			err := plugin.WatchGroup(spec)

			// TODO(chungers) -- yes this is clunky
			if err != nil && strings.Contains(err.Error(), "Already watching") {

				log.Infoln("Already WATCHING", spec.ID, "no action")
				return nil
			}

			if err != nil {
				log.Warningln("Error watching group:", spec.ID, "Err=", err)
			}
			return nil
		})
}

func (m *manager) doUnwatchGroups(config GlobalSpec) error {
	err := m.launchPlugins(config)
	if err != nil {
		return err
	}
	log.Infoln("Unwatching groups")
	return m.execPlugins(config,
		func(plugin group.Plugin, spec group.Spec) error {

			log.Infoln("UNWATCH group", spec.ID, "with spec:", spec)
			err := plugin.UnwatchGroup(spec.ID)
			if err != nil {
				log.Warningln("Error unwatching group:", spec.ID, "Err=", err)
			}
			return nil
		})
}

func (m *manager) launchPlugins(config GlobalSpec) error {
	// Now install / start plugins if not running
	for _, name := range config.findPlugins() {

		endpoint, err := m.plugins.Find(name)
		if err == nil {
			log.Infoln("Found plugin", name, "at", endpoint)
			continue
		}

		// TODO(chungers) - make this async so we can launch them in parallel since there
		// are no dependencies on plugin start up ordering.

		// TODO(chungers) - need to figure out the mechanism for passing in args... like docker run
		log.Infoln("Launching plugin", name)
		running, err := m.launcher.Launch(name)
		if err != nil {
			log.Warningln("Cannot launch plugin", name)
			continue
		}

		err = <-running
		if err == nil {
			log.Infoln("Started plugin", name)
		} else {
			log.Warningln("Failed to start plugin", name, "err=", err)
		}
	}
	return nil
}

func (m *manager) ensurePluginsRunning(config GlobalSpec) error {
	tick := time.Tick(1 * time.Second)
	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for plugins to run")

		case <-tick:
			running, err := m.plugins.List()
			if err != nil {
				return err
			}

			needed := config.findPlugins()
			// TODO(chungers) -- do set intersection
			if len(running) >= len(needed) {
				return nil
			}
		}
	}
}

func (m *manager) execPlugins(config GlobalSpec, work func(group.Plugin, group.Spec) error) error {

	// Do not execute unless all plugins are running.  The entire config should have everything
	// ready as a whole.
	if err := m.ensurePluginsRunning(config); err != nil {
		return err
	}

	running, err := m.plugins.List()
	if err != nil {
		return err
	}

	for id, pluginSpec := range config.Groups {

		log.Infoln("Processing group", id, "with plugin", pluginSpec.Plugin)
		name := pluginSpec.Plugin

		ep, has := running[name]
		if !has {
			log.Warningln("Plugin", name, "isn't running")
			return err
		}

		gp, err := rpc.NewClient(ep.Protocol, ep.Address)
		if err != nil {
			log.Warningln("Cannot contact group", id, " at plugin", name, "endpoint=", ep.Address)
			return err
		}

		log.Debugln("exec on group", id, "plugin=", name)

		// spec is store in the properties
		if pluginSpec.Properties == nil {
			return fmt.Errorf("no spec for group", id, "plugin=", name)
		}

		spec := group.Spec{}
		err = json.Unmarshal([]byte(*pluginSpec.Properties), &spec)
		if err != nil {
			return err
		}

		err = work(gp, spec)
		if err != nil {
			log.Warningln("Error from exec on plugin", err)
			return err
		}

	}

	return nil
}

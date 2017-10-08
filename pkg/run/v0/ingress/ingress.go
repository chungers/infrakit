package ingress

import (
	"github.com/docker/infrakit/pkg/controller/ingress"
	ingress_types "github.com/docker/infrakit/pkg/controller/ingress/types"
	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/launch/inproc"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/manager"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/rpc/client"
	manager_rpc "github.com/docker/infrakit/pkg/rpc/manager"
	"github.com/docker/infrakit/pkg/run"
	"github.com/docker/infrakit/pkg/types"

	// load the handlers for ingress con
	_ "github.com/docker/infrakit/pkg/controller/ingress/swarm"
)

const (
	// Kind is the canonical name of the plugin for starting up, etc.
	Kind = "ingress"

	//EnvOptionsGroup is the env variable to set the backend group controller
	EnvOptionsGroup = "INFRAKIT_INGRESS_GROUP"

	// EnvOptionsBackend is the environment variable to use to set the default value of Options.Backend
	EnvOptionsBackend = "INFRAKIT_INGRESS_BACKEND"
)

var (
	log = logutil.New("module", "run/v0/ingress")
)

func init() {
	inproc.Register(Kind, Run, DefaultOptions)
}

func leadership(plugins func() discovery.Plugins) (manager.Leadership, error) {
	// Scan for a manager
	pm, err := plugins().List()
	if err != nil {
		return nil, err
	}

	for _, endpoint := range pm {
		rpcClient, err := client.New(endpoint.Address, manager.InterfaceSpec)
		if err == nil {
			return manager_rpc.Adapt(rpcClient), nil
		}
	}
	return nil, nil
}

// DefaultOptions container options for default behavior
var DefaultOptions = ingress_types.Options{
	SyncInterval: types.FromDuration(ingress_types.DefaultSyncInterval),
}

// Run runs the plugin, blocking the current thread.  Error is returned immediately
// if the plugin cannot be started.
func Run(plugins func() discovery.Plugins, name plugin.Name,
	config *types.Any) (transport plugin.Transport, impls map[run.PluginCode]interface{}, onStop func(), err error) {

	if plugins == nil {
		panic("no plugins()")
	}

	options := DefaultOptions
	err = config.Decode(&options)
	if err != nil {
		return
	}

	log.Info("Decoded input", "config", options)

	leadership, err := leadership(plugins)
	if err != nil {
		return
	}

	transport.Name = name
	impls = map[run.PluginCode]interface{}{
		run.Controller: ingress.NewTypedControllers(plugins, leadership),
	}

	return
}

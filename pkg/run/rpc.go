package run

import (
	"fmt"

	"github.com/docker/infrakit/pkg/cli"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/manager"
	"github.com/docker/infrakit/pkg/plugin"
	event_rpc "github.com/docker/infrakit/pkg/rpc/event"
	flavor_rpc "github.com/docker/infrakit/pkg/rpc/flavor"
	group_rpc "github.com/docker/infrakit/pkg/rpc/group"
	instance_rpc "github.com/docker/infrakit/pkg/rpc/instance"
	manager_rpc "github.com/docker/infrakit/pkg/rpc/manager"
	metadata_rpc "github.com/docker/infrakit/pkg/rpc/metadata"
	resource_rpc "github.com/docker/infrakit/pkg/rpc/resource"
	"github.com/docker/infrakit/pkg/rpc/server"
	"github.com/docker/infrakit/pkg/spi/event"
	"github.com/docker/infrakit/pkg/spi/flavor"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/spi/metadata"
	"github.com/docker/infrakit/pkg/spi/resource"
)

var log = logutil.New("module", "run")

// PluginCode is the type code for exposing the correct RPC interface for a given object.
// We need type information because some object like Manager implements multiple spi interfaces
// and type information is necessary to know which RPC interface needs to bind to the object.
// This is so that an object that implements both Group and Metadata spi can be bound to separate RPC interfaces.
type PluginCode int

const (
	//Instance is the type code for Instance SPI implementation
	Instance PluginCode = iota
	// Flavor is the type code for Flavor SPI implementation
	Flavor
	// Group is the type code for Group SPI implementation
	Group
	// Manager is the type code for Manager
	Manager
	// Metadata is the type code for Metadata SPI implementation
	Metadata
	// MetadataUpdatable is the type code for updatable Metadata SPI implementation
	MetadataUpdatable
	// Event is the type code for Event SPI implementation
	Event
	// Resource is the type code for Resource SPI implementation
	Resource
)

// ServeRPC starts the RPC endpoint / server given a plugin name for lookup and a list of plugin objects
// that implements the pkg/spi/ interfaces. onStop is a callback invoked when the the endpoint shuts down.
func ServeRPC(name plugin.Name, onStop func(),
	impls map[PluginCode]interface{}) (stoppable server.Stoppable, running <-chan struct{}, err error) {

	// Get the server interfaces to be exported.  Do this by checking on the types of the implementations
	// and wrap the implementation with a rpc adaptor
	plugins := []server.VersionedInterface{}

	for code, p := range impls {

		switch code {

		case Manager:
			log.Debug("manager_rpc.PluginServer", "p", p)
			plugins = append(plugins, manager_rpc.PluginServer(p.(manager.Manager)))
		case Group:
			log.Debug("group_rpc.PluginServer", "p", p)
			plugins = append(plugins, group_rpc.PluginServer(p.(group.Plugin)))
		case Instance:
			log.Debug("instance_rpc.PluginServer", "p", p)
			plugins = append(plugins, instance_rpc.PluginServer(p.(instance.Plugin)))
		case Flavor:
			log.Debug("flavor_rpc.PluginServer", "p", p)
			plugins = append(plugins, flavor_rpc.PluginServer(p.(flavor.Plugin)))
		case MetadataUpdatable:
			log.Debug("metadata_rpc.UpdatablePluginServer", "p", p)
			plugins = append(plugins, metadata_rpc.UpdatablePluginServer(p.(metadata.Updatable)))
		case Metadata:
			log.Debug("metadata_rpc.PluginServer", "p", p)
			plugins = append(plugins, metadata_rpc.PluginServer(p.(metadata.Plugin)))
		case Event:
			log.Debug("event_rpc.PluginServer", "p", p)
			plugins = append(plugins, event_rpc.PluginServer(p.(event.Plugin)))
		case Resource:
			log.Debug("resource_rpc.PluginServer", "p", p)
			plugins = append(plugins, resource_rpc.PluginServer(p.(resource.Plugin)))

		default:
			err = fmt.Errorf("unknown plugin %v, code %v", p, code)
			return
		}

	}

	lookupName, _ := name.GetLookupAndType() // for aws/ec2, start with 'aws' for example.
	stoppable, running = cli.BackgroundPlugin(lookupName, onStop, plugins[0], plugins[1:]...)

	return
}

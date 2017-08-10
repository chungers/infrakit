package group

import (
	"time"

	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/launch/inproc"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/plugin"
	"github.com/docker/infrakit/pkg/plugin/group"
	metadata_plugin "github.com/docker/infrakit/pkg/plugin/metadata"
	flavor_client "github.com/docker/infrakit/pkg/rpc/flavor"
	instance_client "github.com/docker/infrakit/pkg/rpc/instance"
	"github.com/docker/infrakit/pkg/run"
	"github.com/docker/infrakit/pkg/spi/flavor"
	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/types"
)

const (
	// CanonicalName is the canonical name of the plugin. Used for command line identification
	CanonicalName = "group"

	// LookupName is the name used to look up the object via discovery
	LookupName = "group-stateless"

	// EnvOptionsBackend is the environment variable to use to set the default value of Options.Backend
	EnvOptionsBackend = "INFRAKIT_MANAGER_OPTIONS_BACKEND"
)

var log = logutil.New("module", "run/group")

func init() {
	inproc.Register(CanonicalName, Run, DefaultOptions)
}

// Options capture the options for starting up the group controller.
type Options struct {

	// Name is the name to advertise for discovery
	Name plugin.Name

	// PollInterval is the frequency for syncing the state
	PollInterval time.Duration

	// MaxParallelNum is the max number of parallel instance operation. Default =0 (no limit)
	MaxParallelNum uint

	// PollIntervalGroupSpec polls for group spec at this interval to update the metadata paths
	PollIntervalGroupSpec time.Duration

	// PollIntervalGroupDetail polls for group details at this interval to update the metadata paths
	PollIntervalGroupDetail time.Duration
}

// DefaultOptions return an Options with default values filled in.
var DefaultOptions = Options{
	Name:                    plugin.Name(LookupName),
	PollInterval:            10 * time.Second,
	MaxParallelNum:          0,
	PollIntervalGroupSpec:   1 * time.Second,
	PollIntervalGroupDetail: 30 * time.Second,
}

// Run runs the plugin, blocking the current thread.  Error is returned immediately
// if the plugin cannot be started.
func Run(plugins func() discovery.Plugins,
	config *types.Any) (transport plugin.Transport, impls map[run.PluginCode]interface{}, onStop func(), err error) {

	options := DefaultOptions
	err = config.Decode(&options)
	if err != nil {
		return
	}

	instancePluginLookup := func(n plugin.Name) (instance.Plugin, error) {
		endpoint, err := plugins().Find(n)
		if err != nil {
			return nil, err
		}
		return instance_client.NewClient(n, endpoint.Address)
	}

	flavorPluginLookup := func(n plugin.Name) (flavor.Plugin, error) {
		endpoint, err := plugins().Find(n)
		if err != nil {
			return nil, err
		}
		return flavor_client.NewClient(n, endpoint.Address)
	}

	groupPlugin := group.NewGroupPlugin(instancePluginLookup, flavorPluginLookup,
		options.PollInterval, options.MaxParallelNum)

	// Start a poller to load the snapshot and make that available as metadata
	updateSnapshot := make(chan func(map[string]interface{}))
	stopSnapshot := make(chan struct{})
	go func() {
		tick := time.Tick(options.PollIntervalGroupSpec)
		tick2 := time.Tick(options.PollIntervalGroupDetail)
		for {
			select {
			case <-tick:
				// load the specs for the groups
				snapshot := map[string]interface{}{}
				if specs, err := groupPlugin.InspectGroups(); err == nil {
					for _, spec := range specs {
						snapshot[string(spec.ID)] = spec
					}
				} else {
					snapshot["err"] = err
				}

				updateSnapshot <- func(view map[string]interface{}) {
					types.Put([]string{"specs"}, snapshot, view)
				}

			case <-tick2:
				snapshot := map[string]interface{}{}
				// describe the groups and expose info as metadata
				if specs, err := groupPlugin.InspectGroups(); err == nil {
					for _, spec := range specs {
						if description, err := groupPlugin.DescribeGroup(spec.ID); err == nil {
							snapshot[string(spec.ID)] = description
						} else {
							snapshot[string(spec.ID)] = err
						}
					}
				} else {
					snapshot["err"] = err
				}

				updateSnapshot <- func(view map[string]interface{}) {
					types.Put([]string{"groups"}, snapshot, view)
				}

			case <-stopSnapshot:
				log.Info("Snapshot updater stopped")
				return
			}
		}
	}()

	transport.Name = options.Name
	impls = map[run.PluginCode]interface{}{
		run.Metadata: metadata_plugin.NewPluginFromChannel(updateSnapshot),
		run.Group:    groupPlugin,
	}
	onStop = func() {
		close(stopSnapshot)
	}
	return
}

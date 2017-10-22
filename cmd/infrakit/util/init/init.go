package init

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/docker/infrakit/pkg/core"
	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/launch"
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/plugin"
	group_types "github.com/docker/infrakit/pkg/plugin/group/types"
	flavor_rpc "github.com/docker/infrakit/pkg/rpc/flavor"
	metadata_rpc "github.com/docker/infrakit/pkg/rpc/metadata"
	"github.com/docker/infrakit/pkg/run/manager"
	"github.com/docker/infrakit/pkg/run/scope"
	"github.com/docker/infrakit/pkg/run/scope/local"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/spi/metadata"
	"github.com/docker/infrakit/pkg/types"
	"github.com/spf13/cobra"
)

var log = logutil.New("module", "util/init")

func getPluginManager(plugins func() discovery.Plugins,
	services *cli.Services, configURL string) (*manager.Manager, error) {

	parsedRules := []launch.Rule{}

	if configURL != "" {
		buff, err := services.ProcessTemplate(configURL)
		if err != nil {
			return nil, err
		}
		view, err := services.ToJSON([]byte(buff))
		if err != nil {
			return nil, err
		}
		configs := types.AnyBytes(view)
		err = configs.Decode(&parsedRules)
		if err != nil {
			return nil, err
		}
	}
	return manager.ManagePlugins(parsedRules, plugins, true, 5*time.Second)
}

// Command returns the cobra command
func Command(plugins func() discovery.Plugins) *cobra.Command {

	services := cli.NewServices(plugins)

	cmd := &cobra.Command{
		Use:   "init <groups template URL | - >",
		Short: "Generates the init script",
	}

	cmd.Flags().AddFlagSet(services.ProcessTemplateFlags)
	groupID := cmd.Flags().String("group-id", "", "Group ID")
	sequence := cmd.Flags().Uint("sequence", 0, "Sequence in the group")

	configURL := cmd.Flags().String("config-url", "", "URL for the startup configs")

	debug := cmd.Flags().Bool("debug", false, "True to debug with lots of traces")
	waitDuration := cmd.Flags().String("wait", "1s", "Wait for plugins to be ready")

	starts := cmd.Flags().StringSlice("start", []string{}, "start spec for plugin just like infrakit plugin start")

	persist := cmd.Flags().Bool("persist", false, "True to persist any vars into backend")

	cmd.RunE = func(c *cobra.Command, args []string) error {

		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		if !*debug {
			logutil.Configure(&logutil.Options{
				Level:    3,
				Stdout:   false,
				Format:   "term",
				CallFunc: true,
			})
		}

		wait := types.MustParseDuration(*waitDuration)

		pluginManager, err := getPluginManager(plugins, services, *configURL)
		if err != nil {
			return err
		}

		log.Info("Starting up base plugins")
		basePlugins := []string{"vars"}
		if *persist {
			basePlugins = []string{"vars", "manager", "group"} // manager aliased to vars
		}
		for _, base := range basePlugins {
			execName, kind, name, _ := local.StartPlugin(base).Parse()
			err := pluginManager.Launch(execName, kind, name, nil)
			if err != nil {
				log.Error("cannot start base plugin", "spec", base)
				return err
			}
		}
		pluginManager.WaitStarting()
		<-time.After(wait.Duration())

		log.Info("Parsing the input groups.json as template")
		input, err := services.ReadFromStdinIfElse(
			func() bool { return args[0] == "-" },
			func() (string, error) { return services.ProcessTemplate(args[0]) },
			services.ToJSON,
		)
		if err != nil {
			log.Error("processing input", "err", err)
			return err
		}

		// TODO - update the schema soon. This is the Plugin/Properties schema
		type spec struct {
			Plugin     plugin.Name
			Properties struct {
				ID         group.ID
				Properties group_types.Spec
			}
		}

		specs := []spec{}
		err = types.AnyString(input).Decode(&specs)
		if err != nil {
			return err
		}

		var groupSpec *group_types.Spec
		for _, s := range specs {
			if string(s.Properties.ID) == *groupID {
				copy := s.Properties.Properties
				groupSpec = &copy
				break
			}
		}

		if groupSpec == nil {
			return fmt.Errorf("no such group: %v", *groupID)
		}

		// Found group spec
		log.Info("Found group spec", "group", *groupID)

		// Now load the plugins
		pluginsToStart := func() (targets []local.StartPlugin, err error) {
			for _, start := range *starts {
				targets = append(targets, local.StartPlugin(start))
			}

			// TODO -- get the dependencies too
			targets = append(targets, local.FromAddressable(core.AddressableFromPluginName(groupSpec.Flavor.Plugin)))
			return
		}

		buildInit := func(scope scope.Scope) error {

			// Get the flavor properties and use that to call the prepare of the Flavor to generate the init
			endpoint, err := scope.Plugins().Find(groupSpec.Flavor.Plugin)
			if err != nil {
				log.Error("error looking up plugin", "plugin", groupSpec.Flavor.Plugin, "err", err)
				return err
			}

			flavorPlugin, err := flavor_rpc.NewClient(groupSpec.Flavor.Plugin, endpoint.Address)
			if err != nil {
				return err
			}

			cli.MustNotNil(flavorPlugin, "flavor plugin not found", "name", groupSpec.Flavor.Plugin.String())

			instanceSpec := instance.Spec{}
			if lidLen := len(groupSpec.Allocation.LogicalIDs); lidLen > 0 {

				if int(*sequence) >= lidLen {
					return fmt.Errorf("out of bound sequence index: %v in %v", *sequence, groupSpec.Allocation.LogicalIDs)
				}

				lid := instance.LogicalID(groupSpec.Allocation.LogicalIDs[*sequence])
				instanceSpec.LogicalID = &lid
			}

			instanceSpec, err = flavorPlugin.Prepare(groupSpec.Flavor.Properties, instanceSpec,
				groupSpec.Allocation,
				group_types.Index{Group: group.ID(*groupID), Sequence: *sequence})

			if err != nil {
				log.Error("error preparing", "err", err, "spec", instanceSpec)
				return err
			}

			log.Info("apply init template", "init", instanceSpec.Init)

			// Here the Init may contain template vars since in the evaluation of the manager / worker
			// init templates, we do not propapage the vars set in the command line here.
			// So we need to evaluate the entire Init as a template again.
			// TODO - this is really better addressed via some formal globally available var store/section
			// that is always available to the templates at the schema / document level.
			applied, err := services.ProcessTemplate("str://" + instanceSpec.Init)
			if err != nil {
				return err
			}

			if *persist {
				vars := plugin.Name("group/vars")
				log.Info("Persisting data into the backend")
				endpoint, err := scope.Plugins().Find(vars)
				if err != nil {
					return err
				}
				m, err := metadata_rpc.NewClient(vars, endpoint.Address)
				if err != nil {
					return err
				}
				u, is := m.(metadata.Updatable)
				if !is {
					return fmt.Errorf("not updatable")
				}
				_, proposed, cas, err := u.Changes([]metadata.Change{})
				if err != nil {
					return err
				}
				err = u.Commit(proposed, cas)
				log.Info("Committed to vars", "err", err)
				if err != nil {
					return err
				}
			}

			fmt.Print(applied)

			return nil
		}

		return local.Execute(plugins, pluginManager,
			pluginsToStart,
			buildInit,
			local.Options{
				StartWait: wait,
				StopWait:  wait,
			},
		)

	}

	return cmd
}

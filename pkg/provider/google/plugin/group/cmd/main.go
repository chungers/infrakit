package main

import (
	"os"

	"github.com/docker/infrakit.gcp/plugin/group"
	"github.com/docker/infrakit/pkg/cli"
	"github.com/docker/infrakit/pkg/discovery/local"
	"github.com/docker/infrakit/pkg/plugin"
	flavor_client "github.com/docker/infrakit/pkg/rpc/flavor"
	group_plugin "github.com/docker/infrakit/pkg/rpc/group"
	"github.com/docker/infrakit/pkg/spi/flavor"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "GCE group plugin",
	}

	name := cmd.Flags().String("name", "group-gcp", "Plugin name to advertise for discovery")
	logLevel := cmd.Flags().Int("log", cli.DefaultLogLevel, "Logging level. 0 is least verbose. Max is 5")
	project := cmd.Flags().String("project", "", "Google Cloud project")
	zone := cmd.Flags().String("zone", "", "Google Cloud zone")

	cmd.RunE = func(c *cobra.Command, args []string) error {
		cli.SetLogLevel(*logLevel)

		plugins, err := local.NewPluginDiscovery()
		if err != nil {
			return err
		}

		flavorPluginLookup := func(n plugin.Name) (flavor.Plugin, error) {
			endpoint, err := plugins.Find(n)
			if err != nil {
				return nil, err
			}
			return flavor_client.NewClient(n, endpoint.Address)
		}

		cli.RunPlugin(*name, group_plugin.PluginServer(group.NewGCEGroupPlugin(*project, *zone, flavorPluginLookup)))

		return nil
	}

	cmd.AddCommand(cli.VersionCommand())

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

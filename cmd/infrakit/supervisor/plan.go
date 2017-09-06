package manager

import (
	"fmt"
	"os"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/docker/infrakit/pkg/discovery"
	"github.com/docker/infrakit/pkg/types"
	"github.com/spf13/cobra"
)

// Plan command
func Plan(plugins func() discovery.Plugins) *cobra.Command {

	services := cli.NewServices(plugins)

	cmd := &cobra.Command{
		Use:   "plan <url>",
		Short: "Plan returns the plan for this given changes",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 1 {
				cmd.Usage()
				os.Exit(1)
			}

			mgr, err := getManager(plugins)
			if err != nil {
				return err
			}
			cli.MustNotNil(mgr, "manager not found")

			view, err := services.ReadFromStdinIfElse(
				func() bool { return args[0] == "-" },
				func() (string, error) { return services.ProcessTemplate(args[0]) },
				services.ToJSON,
			)
			if err != nil {
				return err
			}

			specs, err := types.SpecsFromString(view)
			if err != nil {
				return err
			}

			// TODO - show diffs of changes
			changes, err := mgr.Plan(specs)
			if err != nil {
				return err
			}

			buff, err := types.AnyValueMust(changes).MarshalYAML()
			if err != nil {
				return err
			}
			fmt.Println(string(buff))

			return nil
		},
	}

	cmd.Flags().AddFlagSet(services.ProcessTemplateFlags)
	return cmd
}

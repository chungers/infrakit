package manager

import (
	"os"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/spf13/cobra"
)

// Terminate returns the terminate command
func Terminate(name string, services *cli.Services) *cobra.Command {
	terminate := &cobra.Command{
		Use:   "terminate",
		Short: "Terminate the cluster (part of Manager interface)",
	}

	terminate.RunE = func(cmd *cobra.Command, args []string) error {

		if len(args) != 0 {
			cmd.Usage()
			os.Exit(1)
		}

		manager, err := Load(services.Plugins())
		if err != nil {
			return nil
		}
		cli.MustNotNil(manager, "manager not found")

		return err
	}
	return terminate
}

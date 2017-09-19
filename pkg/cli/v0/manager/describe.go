package manager

import (
	"os"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/spf13/cobra"
)

// Describe returns the describe command
func Describe(name string, services *cli.Services) *cobra.Command {
	describe := &cobra.Command{
		Use:   "describe",
		Short: "Describe cluster specs (part of Manager interface)",
	}

	describe.RunE = func(cmd *cobra.Command, args []string) error {

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
	return describe
}

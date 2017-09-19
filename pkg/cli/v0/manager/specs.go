package manager

import (
	"os"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/spf13/cobra"
)

// Specs returns the specs command
func Specs(name string, services *cli.Services) *cobra.Command {
	specs := &cobra.Command{
		Use:   "specs",
		Short: "Shows specs information (part of Manager interface)",
	}

	specs.RunE = func(cmd *cobra.Command, args []string) error {

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
	return specs
}

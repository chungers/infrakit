package manager

import (
	"os"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/spf13/cobra"
)

// Enforce returns the enforce command
func Enforce(name string, services *cli.Services) *cobra.Command {
	enforce := &cobra.Command{
		Use:   "enforce",
		Short: "Enforce specs (part of Manager interface)",
	}

	enforce.RunE = func(cmd *cobra.Command, args []string) error {

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
	return enforce
}

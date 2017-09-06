package manager

import (
	"fmt"
	"os"

	"github.com/docker/infrakit/pkg/discovery"
	"github.com/spf13/cobra"
)

// Supervising command
func Supervising(plugins func() discovery.Plugins) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "supervising",
		Short: "Supervising returns the objects the manager is providing leader detection and persistence services",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 0 {
				cmd.Usage()
				os.Exit(1)
			}

			mgr, err := getManager(plugins)
			if err != nil {
				return err
			}

			list, err := mgr.Supervising()
			if err != nil {
				return err
			}

			fmt.Printf("%-15s  %-25s  %-30s\n", "KIND", "NAME", "INTERFACE")

			for _, s := range list {

				fmt.Printf("%-15v  %-25v  %-30v\n", s.Kind, s.Name, s.InterfaceSpec.Encode())
			}
			return nil
		},
	}
	return cmd
}

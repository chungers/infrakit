package main

import (
	"os"

	"github.com/docker/infrakit/discovery"
	"github.com/docker/infrakit/manager"
	manager_rpc "github.com/docker/infrakit/rpc/manager"
	"github.com/spf13/cobra"
)

func managerCommand(plugins func() discovery.Plugins) *cobra.Command {

	name := "manager"
	var managerClient manager.Manager

	cmd := &cobra.Command{
		Use:   "manager",
		Short: "Access manager",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {

			endpoint, err := plugins().Find(name)
			if err != nil {
				return err
			}

			managerClient, err = manager_rpc.NewClient(endpoint.Protocol, endpoint.Address)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&name, "name", name, "Name of manager")

	commit := &cobra.Command{
		Use:   "commit",
		Short: "commit global configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			assertNotNil("no manager", managerClient)

			if len(args) != 1 {
				cmd.Usage()
				os.Exit(1)
			}
			return managerClient.Commit()
		},
	}

	cmd.AddCommand(commit)

	return cmd
}

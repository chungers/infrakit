package main

import (
	"fmt"
	"path/filepath"

	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/infrakit/discovery"
	"github.com/docker/infrakit/manager"
	manager_rpc "github.com/docker/infrakit/rpc/manager"
	"github.com/docker/infrakit/store"
	file_store "github.com/docker/infrakit/store/file"
	swarm_store "github.com/docker/infrakit/store/swarm"
	"github.com/docker/infrakit/util/docker"
	"github.com/spf13/cobra"
)

func managerCommand(plugins func() discovery.Plugins) *cobra.Command {

	name := "group"
	var managerClient manager.Group

	cmd := &cobra.Command{
		Use:   "manager",
		Short: "Access manager",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {

			if c.Use == "commit" {

				// Only commit requires the manager
				endpoint, err := plugins().Find(name)
				if err != nil {
					return err
				}

				managerClient, err = manager_rpc.NewClient(endpoint.Protocol, endpoint.Address)
				if err != nil {
					return err
				}
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
			return managerClient.Commit()
		},
	}

	cmd.AddCommand(commit)

	/////////////////////////////////////////
	// swarm

	// local repo directory
	storeDir := file_store.DefaultStoreDir()
	// remote -- like git remote
	var remote store.Snapshot

	tlsOptions := tlsconfig.Options{}
	host := "unix:///var/run/docker.sock"
	swarm := &cobra.Command{
		Use:   "swarm",
		Short: "swarm mode -- share data in swarm raft store",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			dockerClient, err := docker.NewDockerClient(host, &tlsOptions)
			if err != nil {
				return err
			}

			snapshot, err := swarm_store.NewSnapshot(dockerClient)
			if err != nil {
				return err
			}
			remote = snapshot
			return nil
		},
	}
	swarm.PersistentFlags().StringVar(&host, "host", host, "Docker host")
	swarm.PersistentFlags().StringVar(&tlsOptions.CAFile, "tlscacert", "", "TLS CA cert file path")
	swarm.PersistentFlags().StringVar(&tlsOptions.CertFile, "tlscert", "", "TLS cert file path")
	swarm.PersistentFlags().StringVar(&tlsOptions.KeyFile, "tlskey", "", "TLS key file path")
	swarm.PersistentFlags().BoolVar(&tlsOptions.InsecureSkipVerify, "tlsverify", true, "True to skip TLS")

	pull := &cobra.Command{
		Use:   "pull",
		Short: "pull config from source",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("pulling")

			// local
			local, err := file_store.NewSnapshot(storeDir, "global.config")
			if err != nil {
				return err
			}

			assertNotNil("no remote", remote)

			config := map[string]interface{}{}
			err = remote.Load(&config)
			if err != nil {
				return err
			}

			return local.Save(config)
		},
	}
	pull.Flags().StringVar(&storeDir, "store-dir", storeDir, "Dir to store the config")

	push := &cobra.Command{
		Use:   "push",
		Short: "push config to target",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("pushing")

			// local
			local, err := file_store.NewSnapshot(storeDir, "global.config")
			if err != nil {
				return err
			}

			assertNotNil("no remote", remote)

			config := map[string]interface{}{}
			err = local.Load(&config)
			if err != nil {
				return err
			}

			return remote.Save(config)
		},
	}
	push.Flags().StringVar(&storeDir, "store-dir", storeDir, "Dir to store the config")

	config := &cobra.Command{
		Use:   "config-path",
		Short: "echoes the config file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(filepath.Join(storeDir, "global.config"))
			return nil
		},
	}

	swarm.AddCommand(pull, push, config)

	cmd.AddCommand(swarm)

	return cmd
}

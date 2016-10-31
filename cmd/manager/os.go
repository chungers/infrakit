package main

import (
	"time"

	"github.com/docker/infrakit/discovery"
	os_launcher "github.com/docker/infrakit/launch/os"
	file_leader "github.com/docker/infrakit/leader/file"
	file_store "github.com/docker/infrakit/store/file"
	"github.com/spf13/cobra"
)

func osEnvironment(backend *backend) *cobra.Command {

	pollInterval := 5 * time.Second
	filename := file_leader.DefaultLeaderFile()
	storeDir := file_store.DefaultStoreDir()

	cmd := &cobra.Command{
		Use:   "os",
		Short: "os",
		RunE: func(c *cobra.Command, args []string) error {

			plugins, err := discovery.NewPluginDiscovery()
			if err != nil {
				return err
			}

			leader, err := file_leader.NewDetector(pollInterval, filename, backend.id)
			if err != nil {
				return err
			}

			launcher, err := os_launcher.NewLauncher()
			if err != nil {
				return err
			}

			snapshot, err := file_store.NewSnapshot(storeDir, "global.config")
			if err != nil {
				return err
			}

			backend.plugins = plugins
			backend.leader = leader
			backend.launcher = launcher
			backend.snapshot = snapshot
			return nil
		},
	}
	cmd.Flags().StringVar(&filename, "leader-file", filename, "File used for leader election/detection")
	cmd.Flags().StringVar(&storeDir, "store-dir", storeDir, "Dir to store the config")
	cmd.Flags().DurationVar(&pollInterval, "poll-interval", pollInterval, "Leader polling interval")
	return cmd
}

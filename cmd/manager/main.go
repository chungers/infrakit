package main

import (
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/cli"
	"github.com/docker/infrakit/discovery"
	"github.com/docker/infrakit/launch"
	"github.com/docker/infrakit/leader"
	"github.com/docker/infrakit/manager"
	"github.com/docker/infrakit/rpc"
	manager_rpc "github.com/docker/infrakit/rpc/manager"
	"github.com/docker/infrakit/store"
	"github.com/spf13/cobra"
)

type backend struct {
	id       string
	plugins  discovery.Plugins
	leader   leader.Detector
	snapshot store.Snapshot
	launcher launch.Launcher
}

func main() {

	logLevel := cli.DefaultLogLevel

	backend := &backend{
		id: "manager",
	}

	cmd := &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "Manager",
		PersistentPreRun: func(c *cobra.Command, args []string) {
			cli.SetLogLevel(logLevel)
		},
		PersistentPostRunE: func(c *cobra.Command, args []string) error {

			log.Infoln("Starting up manager:", backend)

			manager := manager.NewManager(backend.launcher, backend.plugins,
				backend.leader, backend.snapshot)

			_, err := manager.Start()
			if err != nil {
				return err
			}

			_, stopped, err := rpc.StartPluginAtPath(
				filepath.Join(backend.plugins.String(), backend.id),
				manager_rpc.RPCServer(manager),
				func() error {
					log.Infoln("Stopping manager")
					manager.Stop()
					return nil
				},
			)

			if err != nil {
				log.Error(err)
			}

			<-stopped // block until done
			return nil
		},
	}
	cmd.PersistentFlags().IntVar(&logLevel, "log", logLevel, "Logging level. 0 is least verbose. Max is 5")
	cmd.PersistentFlags().StringVar(&backend.id, "id", backend.id, "ID of the manager")

	cmd.AddCommand(cli.VersionCommand(), osEnvironment(backend), swarmEnvironment(backend))

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

package manager

import (
	"fmt"
	"os"

	"github.com/docker/infrakit/pkg/cli"
	"github.com/spf13/cobra"
)

// Leader returns the leader command
func Leader(name string, services *cli.Services) *cobra.Command {
	leader := &cobra.Command{
		Use:   "leader",
		Short: "Shows leader information (part of Manager interface)",
	}

	leader.RunE = func(cmd *cobra.Command, args []string) error {

		if len(args) != 0 {
			cmd.Usage()
			os.Exit(1)
		}

		manager, err := Load(services.Plugins())
		if err != nil {
			return nil
		}
		cli.MustNotNil(manager, "manager not found")

		isleader := "unknown"
		if l, err := manager.IsLeader(); err == nil {
			if l {
				isleader = "true"
			} else {
				isleader = "false"
			}
		} else {
			log.Warn("error determining leader", "err", err)
		}
		fmt.Printf("IsLeader       : %v\n", isleader)

		location := "unknown"
		if l, err := manager.LeaderLocation(); err == nil {
			location = l.String()
		} else {
			log.Warn("error getting location of leader", "err", err)
		}
		fmt.Printf("LeaderLocation : %v\n", location)
		return err
	}
	return leader
}

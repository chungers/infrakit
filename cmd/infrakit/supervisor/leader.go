package manager

import (
	"fmt"
	"os"

	"github.com/docker/infrakit/pkg/discovery"
	"github.com/spf13/cobra"
)

// Leader command
func Leader(plugins func() discovery.Plugins) *cobra.Command {

	leader := &cobra.Command{
		Use:   "leader",
		Short: "Leader returns the leadership information",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) != 0 {
				cmd.Usage()
				os.Exit(1)
			}

			mgr, err := getManager(plugins)
			if err != nil {
				return err
			}
			isleader := "unknown"
			if l, err := mgr.IsLeader(); err == nil {
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
			if l, err := mgr.LeaderLocation(); err == nil {
				location = l.String()
			} else {
				log.Warn("error getting location of leader", "err", err)
			}
			fmt.Printf("LeaderLocation : %v\n", location)
			return nil
		},
	}
	return leader
}

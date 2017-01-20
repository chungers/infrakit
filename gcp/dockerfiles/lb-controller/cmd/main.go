package main

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/opts"
	"github.com/docker/editions/pkg/loadbalancer"
	"github.com/docker/editions/pkg/loadbalancer/gcp"
	"github.com/spf13/cobra"
)

var (
	defaultHost = "unix://" + opts.DefaultUnixSocket
	logLevel    = len(log.AllLevels) - 1
)

func main() {
	cmd := &cobra.Command{
		Use:   "loadbalancer",
		Short: "Load Balancer Controller for Editions",
		PersistentPreRun: func(_ *cobra.Command, args []string) {
			if logLevel > len(log.AllLevels)-1 {
				logLevel = len(log.AllLevels) - 1
			} else if logLevel < 0 {
				logLevel = 0
			}

			log.SetLevel(log.AllLevels[logLevel])
		},
	}

	cmd.PersistentFlags().IntVar(&logLevel, "log", logLevel, "Logging level. 0 is least verbose. Max is 5")

	cmd.AddCommand(runCommand())

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}

func runCommand() *cobra.Command {
	interval := 3

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the publisher",
		RunE: func(_ *cobra.Command, args []string) error {
			log.Infoln("Connecting to docker:", defaultHost)

			client, err := loadbalancer.NewDockerClient(defaultHost, nil)
			if err != nil {
				return err
			}

			ctx := context.Background()
			leader, err := loadbalancer.AmISwarmLeader(client, ctx)
			if err != nil {
				leader = false
			}

			// Just poll until I become the leader...
			// TODO(chungers) - This doesn't account for the change from leader to
			// non-leader. It's not clear whether this can happen when a leader
			// node gets demoted in the raft quorum.
			if !leader {
				log.Infoln("Not a leader...")
				tick := time.Tick(time.Duration(interval) * time.Second)
			leaderPoll:
				for {
					select {
					case <-ctx.Done():
						return err
					case <-tick:
						leader, _ = loadbalancer.AmISwarmLeader(client, ctx)

						if leader {
							log.Infoln("I am the leader. Now polling for services.")
							break leaderPoll
						} else {
							log.Infoln("Not a leader. Check back later")
						}
					}
				}
			}

			log.Infoln("Running on leader node")

			options := loadbalancer.Options{
				RemoveListeners:   true,
				PublishAllExposed: true,
			}

			actionExposePublishedPorts := loadbalancer.ExposeServicePortInExternalLoadBalancer(
				func() map[string]loadbalancer.Driver {
					hostMapping := map[string]loadbalancer.Driver{}

					host := "default"
					elbName := "default"

					elb, errDriver := gcp.NewLoadBalancerDriver(elbName)
					if errDriver != nil {
						log.Warningln("Cannot create provisioner for", elbName)
					} else {
						hostMapping[host] = elb
						log.Infoln("Located external load balancer", elbName, "for", host)
					}

					return hostMapping
				}, options)

			poller, err := loadbalancer.NewServicePoller(client, time.Duration(interval)*time.Second).
				AddService("elb-rule", loadbalancer.AnyServices, actionExposePublishedPorts).
				Build()
			if err != nil {
				return err
			}

			return poller.Run(ctx)
		},
	}

	cmd.Flags().IntVar(&interval, "poll_interval", interval, "Polling interval in seconds")

	return cmd
}

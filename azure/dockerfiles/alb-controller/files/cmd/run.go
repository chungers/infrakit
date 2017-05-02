package main

import (
	"io/ioutil"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/editions/pkg/loadbalancer"
	"github.com/docker/editions/pkg/loadbalancer/azure"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
)

func runCommand() *cobra.Command {

	albOptions := new(azure.Options)
	elbConfig := ""
	interval := 3
	forceLeader := false
	hardSync := true

	options := loadbalancer.Options{
		RemoveListeners:   true,
		PublishAllExposed: true,
	}

	doHealthCheck := false
	healthCheck := loadbalancer.HealthCheck{
		Port:            0,
		Healthy:         2,
		Unhealthy:       10,
		TimeoutSeconds:  9,
		IntervalSeconds: 10,
	}

	defaultLBName := ""

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the publisher",
		RunE: func(_ *cobra.Command, args []string) error {
			log.Infoln("Connecting to docker:", host)

			ctx := context.Background()

			if doHealthCheck {
				options.HealthCheck = &healthCheck
			}

			tls := &tlsOptions
			if tlsOptions.KeyFile == "" || tlsOptions.CAFile == "" || tlsOptions.CertFile == "" {
				log.Infoln("Not using TLS")
				tls = nil
			}
			client, err := loadbalancer.NewDockerClient(host, tls)
			if err != nil {
				return err
			}

			leader, err := loadbalancer.AmISwarmLeader(client, ctx)
			if err != nil {
				leader = false
			}

			// Just poll until I become the leader...
			// TODO(chungers) - This doesn't account for the change from leader to
			// non-leader. It's not clear whether this can happen when a leader
			// node gets demoted in the raft quorum.
			if !leader && !forceLeader {
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
							log.Infoln("Not a leader.  Check back later")
						}
					}
				}
			}

			log.Infoln("Running on leader node: forceLeader=", forceLeader)

			actionExposePublishedPorts := loadbalancer.ExposeServicePortInExternalLoadBalancer(
				func() map[string]loadbalancer.Driver {
					// Loads the hostname to elb mapping
					buff, err := ioutil.ReadFile(elbConfig)
					if err != nil {
						panic(err)
					}

					log.Infoln("Read config:", string(buff))

					mapping := map[string]string{}

					err = yaml.Unmarshal(buff, &mapping)
					if err != nil {
						log.Warningln("Error parsing config:", err)
					}

					if defaultLBName != "" {
						mapping["default"] = defaultLBName
						log.Infoln("Default LB name override to", defaultLBName)
					}

					log.Infoln("ELB mapping:", mapping)

					// Must return unique instances and not multiple instances with same name
					elbByName := map[string]loadbalancer.Driver{}
					hostMapping := map[string]loadbalancer.Driver{}

					for host, elbName := range mapping {

						var elb loadbalancer.Driver
						if v, has := elbByName[elbName]; !has {

							cred := azure.NewCredential()

							err := cred.Authenticate(*albOptions)
							if err != nil {
								log.Warningln("Cannot authenticate for azure", err)
								continue
							}

							client, err := azure.CreateALBClient(cred, *albOptions)
							if err != nil {
								log.Warningln("Cannot create client", err)
								continue
							}

							v, err = azure.NewLoadBalancerDriver(client, *albOptions, elbName)
							if err != nil {
								log.Warningln("Cannot create provisioner for", elbName)
								continue
							}

							elb = v
							elbByName[elbName] = elb
						} else {
							elb = v
						}

						hostMapping[host] = elb
						log.Infoln("Located external load balancer", elbName, "for", host)

					}
					return hostMapping
				}, options)

			poller, err := loadbalancer.NewServicePoller(client, time.Duration(interval)*time.Second).
				AddService("elb-rule", loadbalancer.AnyServices, actionExposePublishedPorts).
				SetHardSyncWithLB(true).
				Build()

			if err != nil {
				return err
			}

			return poller.Run(ctx)
		},
	}

	cmd.Flags().BoolVar(&forceLeader, "leader", forceLeader, "True forces this instance to be a leader")
	cmd.Flags().BoolVar(&hardSync, "hard_sync", true, "True to force syncing")
	cmd.Flags().IntVar(&interval, "poll_interval", interval, "Polling interval in seconds")
	cmd.Flags().StringVar(&elbConfig, "config", "/var/lib/docker/editions/elb.config", "Loadbalancer config")

	cmd.Flags().IntVar(&albOptions.PollingDelaySeconds, "polling_delay", 5, "Polling delay in seconds")
	cmd.Flags().IntVar(&albOptions.PollingDurationSeconds, "polling_duration", 30, "Polling duration in seconds")
	cmd.Flags().StringVar(&albOptions.Environment, "environment", azure.DefaultEnvironment, "environment")
	cmd.Flags().StringVar(&albOptions.OAuthClientID, "oauth_client_id", "", "OAuth client ID")

	cmd.Flags().StringVar(&albOptions.ADClientID, "ad_app_id", "", "AD app ID")
	cmd.Flags().StringVar(&albOptions.ADClientSecret, "ad_app_secret", "", "AD app secret")
	cmd.Flags().StringVar(&albOptions.SubscriptionID, "subscription_id", "", "subscription ID")
	cmd.Flags().StringVar(&albOptions.ResourceGroupName, "resource_group", "", "resource group name")

	cmd.Flags().StringVar(&defaultLBName, "default_lb_name", defaultLBName, "Set to override the name of the default LB")

	cmd.Flags().BoolVar(&doHealthCheck, "health_check", doHealthCheck,
		"True to enable auto config ELB health check.")
	cmd.Flags().BoolVar(&options.RemoveListeners, "gc", options.RemoveListeners,
		"True to remove listeners in load balancer")
	cmd.Flags().BoolVar(&options.PublishAllExposed, "all", options.PublishAllExposed,
		"True to publish all exposed ports in a service")

	cmd.Flags().IntVar(&healthCheck.Healthy, "health_check_healthy_ct", healthCheck.Healthy,
		"ELB health check count to be considered healthy")
	cmd.Flags().IntVar(&healthCheck.Unhealthy, "health_check_unhealthy_ct", healthCheck.Unhealthy,
		"ELB health check count to be considered unhealthy")
	cmd.Flags().IntVar(&healthCheck.IntervalSeconds, "health_check_pint_interval", healthCheck.IntervalSeconds,
		"ELB health check ping interval seconds")
	cmd.Flags().IntVar(&healthCheck.TimeoutSeconds, "health_check_timeout", healthCheck.TimeoutSeconds,
		"ELB health check timeout seconds")

	return cmd
}

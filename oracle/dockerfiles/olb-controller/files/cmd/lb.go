package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/editions/pkg/loadbalancer"
	"github.com/docker/editions/pkg/loadbalancer/oracle"
	"github.com/spf13/cobra"
)

func olbCommand() *cobra.Command {

	olbOptions := new(oracle.Options)

	cmd := &cobra.Command{
		Use:   "olb",
		Short: "Ops on the OCLB",
	}
	cmd.PersistentFlags().IntVar(&olbOptions.PollingDelaySeconds, "polling_delay", 5, "Polling delay")
	cmd.PersistentFlags().IntVar(&olbOptions.PollingDurationSeconds, "polling_duration", 30, "Polling duration")
	cmd.PersistentFlags().StringVar(&olbOptions.UserID, "user_ocid", "", "OCID of the user calling the API")
	cmd.PersistentFlags().StringVar(&olbOptions.ComponentID, "component_ocid", "", "OCID of the component")
	cmd.PersistentFlags().StringVar(&olbOptions.TenancyID, "tenancy_ocid", "", "OCID of the user tenancy")
	cmd.PersistentFlags().StringVar(&olbOptions.Fingerprint, "fingerprint", "", "Fingerprint for the key pair being used")
	cmd.PersistentFlags().StringVar(&olbOptions.KeyFile, "key_file", "", "Full path and filename of the private key")
	cmd.PersistentFlags().StringVar(&olbOptions.StackName, "stack_name", "", "Name of the deployment")

	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describes the OCLB",
		RunE: func(_ *cobra.Command, args []string) error {
			log.Infoln("Running describe OCLB")
			// LoadBalancer OCID
			name := args[0]

			apiClient, err := oracle.NewClient(olbOptions)
			if err != nil {
				return err
			}
			client := oracle.CreateOLBClient(apiClient, olbOptions)

			p, err := oracle.NewLoadBalancerDriver(client, name, olbOptions)
			if err != nil {
				return err
			}
			result, err := p.Routes()
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		},
	}

	publishOptions := new(struct {
		ExtPort     uint32
		BackendPort uint32
		Protocol    string
		Certificate string
	})

	publishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish a service at given ports",
		RunE: func(_ *cobra.Command, args []string) error {
			// LoadBalancer OCID
			name := args[0]

			apiClient, err := oracle.NewClient(olbOptions)
			if err != nil {
				return err
			}
			client := oracle.CreateOLBClient(apiClient, olbOptions)

			p, err := oracle.NewLoadBalancerDriver(client, name, olbOptions)
			if err != nil {
				return err
			}

			result, err := p.Publish(loadbalancer.Route{
				Port:             publishOptions.BackendPort,
				Protocol:         loadbalancer.ProtocolFromString(publishOptions.Protocol),
				LoadBalancerPort: publishOptions.ExtPort,
				Certificate:      &publishOptions.Certificate,
			})
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		},
	}
	publishCmd.Flags().Uint32Var(&publishOptions.ExtPort, "ext_port", 80, "External port")
	publishCmd.Flags().Uint32Var(&publishOptions.BackendPort, "backend_port", 30000, "Backend port")
	publishCmd.Flags().StringVar(&publishOptions.Protocol, "protocol", "http", "Protocol: http|https|tcp|tls")
	publishCmd.Flags().StringVar(&publishOptions.Certificate, "certificate", "", "AWS ACM certificate ARN")

	unpublishPort := uint32(80)
	unpublishCmd := &cobra.Command{
		Use:   "unpublish",
		Short: "Unpublish a service at given port",
		RunE: func(_ *cobra.Command, args []string) error {
			// LoadBalancer OCID
			name := args[0]

			apiClient, err := oracle.NewClient(olbOptions)
			if err != nil {
				return err
			}
			client := oracle.CreateOLBClient(apiClient, olbOptions)

			p, err := oracle.NewLoadBalancerDriver(client, name, olbOptions)
			if err != nil {
				return err
			}

			result, err := p.Unpublish(unpublishPort)
			if err != nil {
				return err
			}
			fmt.Println(result)
			return nil
		},
	}
	unpublishCmd.Flags().Uint32Var(&unpublishPort, "ext_port", unpublishPort, "External port")

	cmd.AddCommand(describeCmd, publishCmd, unpublishCmd)

	return cmd
}

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// AWSWeb interface for all Azure calls
type AWSWeb struct {
}

// TokenManager obtain the AWS Swarm Manager Token
func (a AWSWeb) TokenManager(w http.ResponseWriter, r *http.Request) {
	// get the swarm manager token, if they are a manager node,
	// and are not already in the swarm. Block otherwise
	RequestInfo(r)
	ip := RequestIP(r)
	inSwarm := alreadyInSwarm(ip)
	isManager := isManagerNode(a, ip)

	if inSwarm || !isManager {
		// they are either already in the swarm, or they are not a manager
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Access Denied")
		return
	}

	// They are not in the swarm, and they are a manager, so good to go.
	cli, ctx := DockerClient()
	swarm, err := cli.SwarmInspect(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
		return
	}

	fmt.Fprintf(w, swarm.JoinTokens.Manager)
}

// TokenWorker obtain the AWS Swarm Worker Token
func (a AWSWeb) TokenWorker(w http.ResponseWriter, r *http.Request) {
	// get the swarm worker token, if they are a worker node,
	// and are not already in the swarm. block otherwise
	RequestInfo(r)

	ip := RequestIP(r)
	inSwarm := alreadyInSwarm(ip)
	isWorker := isWorkerNode(a, ip)

	if inSwarm || !isWorker {
		// they are either already in the swarm, or they are not a worker
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, "Access Denied")
		return
	}

	// They are not in the swarm, and they are a worker, so good to go.
	cli, ctx := DockerClient()
	swarm, err := cli.SwarmInspect(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
		return
	}

	fmt.Fprintf(w, swarm.JoinTokens.Worker)
}

// Managers get list of AWS manager instances
func (a AWSWeb) Managers() []WebInstance {
	// get the instances from AWS Manager security group
	customFilter := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("tag:swarm-node-type"),
			Values: []*string{aws.String("manager")},
		},
	}

	// ec2 classic would be just group-id, VPC would be instance.group-id
	customFilter = append(customFilter, &ec2.Filter{
		Name:   aws.String("instance.group-id"),
		Values: []*string{aws.String(os.Getenv("MANAGER_SECURITY_GROUP_ID"))},
	})

	return awsInstances(customFilter)
}

// Workers get list of AWS worker instances
func (a AWSWeb) Workers() []WebInstance {
	// get the instances from AWS worker security group
	customFilter := []*ec2.Filter{
		&ec2.Filter{
			Name:   aws.String("tag:swarm-node-type"),
			Values: []*string{aws.String("worker")},
		},
	}

	// ec2 classic would be just group-id, VPC would be instance.group-id
	customFilter = append(customFilter, &ec2.Filter{
		Name:   aws.String("instance.group-id"),
		Values: []*string{aws.String(os.Getenv("WORKER_SECURITY_GROUP_ID"))},
	})

	return awsInstances(customFilter)
}

// Instances prints the list of Instance ID and their private IP
func (a AWSWeb) Instances(w http.ResponseWriter, r *http.Request) {
	// show both manager and worker instances
	RequestInfo(r)
	instances := a.Managers()
	for _, instance := range instances {
		fmt.Fprintf(w, "%s\n", instance.PrivateIPAddress)
	}
	instances = a.Workers()
	for _, instance := range instances {
		fmt.Fprintf(w, "%s\n", instance.PrivateIPAddress)
	}
}

// DtrInfo prints the list of Instance ID and their private IP
func (a AWSWeb) DtrInfo(w http.ResponseWriter, r *http.Request) {
	// show DTR image used on manager
	fmt.Fprintf(w, "%s\n", getDTRImage())
}

func awsInstances(customFilters []*ec2.Filter) []WebInstance {
	// get the instances from AWS, takes a filter to limit the results.
	client := ec2.New(session.New(&aws.Config{}))

	// Only grab instances that are running or just started
	filters := []*ec2.Filter{
		{
			Name: aws.String("instance-state-name"),
			Values: []*string{
				aws.String("running"),
				aws.String("pending"),
			},
		},
	}

	for _, custom := range customFilters {
		filters = append(filters, custom)
	}

	request := ec2.DescribeInstancesInput{Filters: filters}
	result, err := client.DescribeInstances(&request)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		logrus.Error(err.Error())
	}

	var instances []WebInstance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			var instanceState, instancePubIP, instanceAZ, instancePrivateIP string
			if instance.State != nil {
				instanceState = *instance.State.Name
			}
			if instance.PublicIpAddress != nil {
				instancePubIP = *instance.PublicIpAddress
			}
			if instance.PrivateIpAddress != nil {
				instancePrivateIP = *instance.PrivateIpAddress
			}
			if instance.Placement != nil {
				instanceAZ = *instance.Placement.AvailabilityZone
			}

			aInstance := WebInstance{
				InstanceID:       *instance.InstanceId,
				InstanceType:     *instance.InstanceType,
				PublicIPAddress:  instancePubIP,
				PrivateIPAddress: instancePrivateIP,
				InstanceState:    instanceState,
				InstanceAZ:       instanceAZ,
			}
			instances = append(instances, aInstance)
		}
	}
	return instances
}

// NewAWSClient returns a struct of the web struct
func NewAWSClient() AWSWeb {
	return AWSWeb{}
}

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AwsInstance struct {
	InstanceID       string
	InstanceType     string
	PublicIPAddress  string
	PrivateIPAddress string
	InstanceState    string
	InstanceAZ       string
}

type AWSWeb struct {
}

func (a AWSWeb) TokenManager(w http.ResponseWriter, r *http.Request) {
	// get the swarm manager token, if they are a manager node,
	// and are not already in the swarm. Block otherwise
	RequestInfo(r)
	found := alreadyInSwarm(r)
	isManager := isManagerNode(r)

	if found || !isManager {
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
	}

	fmt.Fprintf(w, swarm.JoinTokens.Manager)
}

func (a AWSWeb) TokenWorker(w http.ResponseWriter, r *http.Request) {
	// get the swarm worker token, if they are a worker node,
	// and are not already in the swarm. block otherwise
	RequestInfo(r)

	found := alreadyInSwarm(r)
	isWorker := isWorkerNode(r)

	if found || !isWorker {
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
	}

	fmt.Fprintf(w, swarm.JoinTokens.Worker)
}

func alreadyInSwarm(r *http.Request) bool {
	// Is the node making the request, already in the swarm.
	ip := RequestIP(r)
	nodes := SwarmNodes()
	for _, node := range nodes {
		nodeIP := convertAWSHostToIP(node.Description.Hostname)
		if ip == nodeIP {
			return true
		}
	}
	return false
}

func isManagerNode(r *http.Request) bool {
	// Is the node making the request a manager node
	ip := RequestIP(r)
	instances := awsManagers()
	for _, instance := range instances {
		if ip == instance.PrivateIPAddress {
			return true
		}
	}
	return false
}

func isWorkerNode(r *http.Request) bool {
	// Is the node making the request a worker node
	ip := RequestIP(r)
	instances := awsWorkers()
	for _, instance := range instances {
		if ip == instance.PrivateIPAddress {
			return true
		}
	}
	return false
}

func awsWorkers() []AwsInstance {
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

func awsManagers() []AwsInstance {
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

func awsInstances(customFilters []*ec2.Filter) []AwsInstance {
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
		fmt.Println(err.Error())
	}

	var instances []AwsInstance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			awsInstance := AwsInstance{
				InstanceID:       *instance.InstanceId,
				InstanceType:     *instance.InstanceType,
				PublicIPAddress:  *instance.PublicIpAddress,
				PrivateIPAddress: *instance.PrivateIpAddress,
				InstanceState:    *instance.State.Name,
				InstanceAZ:       *instance.Placement.AvailabilityZone,
			}
			instances = append(instances, awsInstance)
		}
	}
	return instances
}

func convertAWSHostToIP(hostStr string) string {
	// This is risky, this assumes the following formation for hosts in swarm node ls
	// ip-10-0-3-149.ec2.internal
	// there was one use case when someone had an old account, and their hostnames were not
	// in this format. they just had
	// ip-192-168-33-67
	// not sure how many other formats there are.
	// This will work for both formats above.
	hostSplit := strings.Split(hostStr, ".")
	host := hostSplit[0]
	host = strings.Replace(host, "ip-", "", -1)
	ip := strings.Replace(host, "-", ".", -1)
	return ip
}

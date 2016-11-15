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

type awsInstance struct {
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
	ip := RequestIP(r)
	inSwarm := alreadyInSwarm(ip)
	isManager := isManagerNode(ip)

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

func (a AWSWeb) TokenWorker(w http.ResponseWriter, r *http.Request) {
	// get the swarm worker token, if they are a worker node,
	// and are not already in the swarm. block otherwise
	RequestInfo(r)

	ip := RequestIP(r)
	inSwarm := alreadyInSwarm(ip)
	isWorker := isWorkerNode(ip)

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

func alreadyInSwarm(ip string) bool {
	// Is the node making the request, already in the swarm.
	nodes := SwarmNodes()
	for _, node := range nodes {
		nodeIP := ConvertAWSHostToIP(node.Description.Hostname)
		if ip == nodeIP {
			return true
		}
	}
	return false
}

func isManagerNode(ip string) bool {
	// Is the node making the request a manager node
	return isNodeInList(ip, awsManagers())
}

func isWorkerNode(ip string) bool {
	// Is the node making the request a worker node
	return isNodeInList(ip, awsWorkers())
}

func isNodeInList(ip string, instances []awsInstance) bool {
	// given an IP, find out if it is in the instance list.
	for _, instance := range instances {
		if ip == instance.PrivateIPAddress {
			return true
		}
	}
	return false
}

func awsWorkers() []awsInstance {
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

func awsManagers() []awsInstance {
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

func awsInstances(customFilters []*ec2.Filter) []awsInstance {
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

	var instances []awsInstance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			aInstance := awsInstance{
				InstanceID:       *instance.InstanceId,
				InstanceType:     *instance.InstanceType,
				PublicIPAddress:  *instance.PublicIpAddress,
				PrivateIPAddress: *instance.PrivateIpAddress,
				InstanceState:    *instance.State.Name,
				InstanceAZ:       *instance.Placement.AvailabilityZone,
			}
			instances = append(instances, aInstance)
		}
	}
	return instances
}

func ConvertAWSHostToIP(hostStr string) string {
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

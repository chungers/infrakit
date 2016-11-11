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
	GetRequestInfo(r)
	found := AlreadyInSwarm(r)
	isManager := IsManagerNode(r)

	if found || !isManager {
		// they are either already in the swarm, or they are not a manager
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Access Denied\n")
		return
	}

	// They are not in the swarm, and they are a manager, so good to go.
	cli, ctx := GetDockerClient()
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
	GetRequestInfo(r)

	found := AlreadyInSwarm(r)
	isWorker := IsWorkerNode(r)

	if found || !isWorker {
		// they are either already in the swarm, or they are not a worker
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Access Denied\n")
		return
	}

	// They are not in the swarm, and they are a worker, so good to go.
	cli, ctx := GetDockerClient()
	swarm, err := cli.SwarmInspect(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
	}

	fmt.Fprintf(w, swarm.JoinTokens.Worker)
}

func (a AWSWeb) CheckManager(w http.ResponseWriter, r *http.Request) {
	// check if this node is a manager and already in the swarm.
	GetRequestInfo(r)
	ip := GetRequestIP(r)
	found := AlreadyInSwarm(r)
	isManager := IsManagerNode(r)
	fmt.Fprintf(w, "IP:%s;InSwarm:%v;IsManager:%v;\n", ip, found, isManager)
}

func (a AWSWeb) CheckWorker(w http.ResponseWriter, r *http.Request) {
	// check if this node is a worker, and already in the swarm.
	GetRequestInfo(r)
	ip := GetRequestIP(r)
	found := AlreadyInSwarm(r)
	isWorker := IsWorkerNode(r)
	fmt.Fprintf(w, "IP:%s;InSwarm:%v;IsWorker:%v;\n", ip, found, isWorker)
}

func (a AWSWeb) Instances(w http.ResponseWriter, r *http.Request) {
	// show both manager and worker instances
	GetRequestInfo(r)
	fmt.Fprintf(w, "Managers: \n")
	instances := GetAWSManagers()
	for _, instance := range instances {
		fmt.Fprintf(w, "%s %s\n", instance.InstanceID, instance.PrivateIPAddress)
	}
	fmt.Fprintf(w, "Workers: \n")
	instances = GetAWSWorkers()
	for _, instance := range instances {
		fmt.Fprintf(w, "%s %s\n", instance.InstanceID, instance.PrivateIPAddress)
	}
}

func (a AWSWeb) ManagerInstances(w http.ResponseWriter, r *http.Request) {
	// only show manager instances
	GetRequestInfo(r)
	instances := GetAWSManagers()
	for _, instance := range instances {
		fmt.Fprintf(w, "%s %s\n", instance.InstanceID, instance.PrivateIPAddress)
	}
}

func (a AWSWeb) WorkerInstances(w http.ResponseWriter, r *http.Request) {
	// only show worker instances
	GetRequestInfo(r)
	instances := GetAWSWorkers()
	for _, instance := range instances {
		fmt.Fprintf(w, "%s %s\n", instance.InstanceID, instance.PrivateIPAddress)
	}
}

func (a AWSWeb) Nodes(w http.ResponseWriter, r *http.Request) {
	// print the list of nodes in the swarm.
	GetRequestInfo(r)
	nodes := GetSwarmNodes()
	for _, node := range nodes {
		ip := ConvertAWSHostToIP(node.Description.Hostname)
		fmt.Fprintf(w, "%s %s %s\n", node.ID[:15], node.Description.Hostname, ip)
	}
}

func AlreadyInSwarm(r *http.Request) bool {
	// Is the node making the request, already in the swarm.
	ip := GetRequestIP(r)
	nodes := GetSwarmNodes()
	for _, node := range nodes {
		nodeIP := ConvertAWSHostToIP(node.Description.Hostname)
		if ip == nodeIP {
			return true
		}
	}
	return false
}

func IsManagerNode(r *http.Request) bool {
	// Is the node making the request a manager node
	ip := GetRequestIP(r)
	instances := GetAWSManagers()
	for _, instance := range instances {
		if ip == instance.PrivateIPAddress {
			return true
		}
	}
	return false
}

func IsWorkerNode(r *http.Request) bool {
	// Is the node making the request a worker node
	ip := GetRequestIP(r)
	instances := GetAWSWorkers()
	for _, instance := range instances {
		if ip == instance.PrivateIPAddress {
			return true
		}
	}
	return false
}

func GetAWSWorkers() []AwsInstance {
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

	return getInstances(customFilter)
}

func GetAWSManagers() []AwsInstance {
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

	return getInstances(customFilter)
}

func getInstances(customFilters []*ec2.Filter) []AwsInstance {
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

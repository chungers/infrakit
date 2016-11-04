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

func AwsTokenManager(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println("Endpoint Hit: tokenManager")
}

func AwsTokenWorker(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println("Endpoint Hit: tokenWorker")
}

func AwsCheckManager(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)
	ip := GetRequestIP(r)
	found := AlreadyInSwarm(r)
	isManager := IsManagerNode(r)
	fmt.Fprintf(w, "IP:%s;In Swarm:%v;Is Manager:%v;\n", ip, found, isManager)
}

func AwsCheckWorker(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)
	ip := GetRequestIP(r)
	found := AlreadyInSwarm(r)
	isWorker := IsWorkerNode(r)
	fmt.Fprintf(w, "IP:%s;In Swarm:%v;Is Worker:%v;\n", ip, found, isWorker)
}

func AwsInstances(w http.ResponseWriter, r *http.Request) {
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

func AwsNodes(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)
	nodes := GetSwarmNodes()
	for _, node := range nodes {
		ip := ConvertAWSHostToIP(node.Description.Hostname)
		fmt.Fprintf(w, "%s %s %s\n", node.ID[:10], node.Description.Hostname, ip)
	}
}

func AlreadyInSwarm(r *http.Request) bool {
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

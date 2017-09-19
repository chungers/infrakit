package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/FrenchBen/oracle-sdk-go/api"
	"github.com/FrenchBen/oracle-sdk-go/bmc"
	"github.com/FrenchBen/oracle-sdk-go/core"
	"github.com/Sirupsen/logrus"
)

// OracleWeb interface for all Azure calls
type OracleWeb struct {
	Client      *core.Client
	StackName   string
	ComponentID string
}

// TokenManager obtain the Oracle Swarm Manager Token
func (o OracleWeb) TokenManager(w http.ResponseWriter, r *http.Request) {
	// get the swarm manager token, if they are a manager node,
	// and are not already in the swarm. Block otherwise
	RequestInfo(r)
	ip := RequestIP(r)
	inSwarm := alreadyInSwarm(ip)
	isManager := isManagerNode(o, ip)

	if inSwarm || !isManager {
		// they are either already in the swarm, or they are not a manager
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Access Denied: Manager: %t - in Swarm: %t", isManager, inSwarm)
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

// TokenWorker obtain the Oracle Swarm Worker Token
func (o OracleWeb) TokenWorker(w http.ResponseWriter, r *http.Request) {
	// get the swarm worker token, if they are a worker node,
	// and are not already in the swarm. block otherwise
	RequestInfo(r)

	ip := RequestIP(r)
	inSwarm := alreadyInSwarm(ip)
	isWorker := isWorkerNode(o, ip)

	if inSwarm || !isWorker {
		// they are either already in the swarm, or they are not a worker
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Access Denied")
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

// Managers get list of Oracle manager instances
func (o OracleWeb) Managers() []WebInstance {
	// get the instances from Oracle Manager security group
	customFilter := core.InstancesParameters{
		Filter: &core.InstanceFilter{
			DisplayName:    fmt.Sprintf("%s-manager-*", o.StackName),
			LifeCycleState: "running",
		},
	}
	return oracleInstances(o.Client, customFilter)
}

// Workers get list of Oracle worker instances
func (o OracleWeb) Workers() []WebInstance {
	// get the instances from Oracle worker security group
	customFilter := core.InstancesParameters{
		Filter: &core.InstanceFilter{
			DisplayName:    fmt.Sprintf("%s-worker-*", o.StackName),
			LifeCycleState: "running",
		},
	}

	return oracleInstances(o.Client, customFilter)
}

// Instances prints the list of Instance ID and their private IP
func (o OracleWeb) Instances(w http.ResponseWriter, r *http.Request) {
	// show both manager and worker instances
	RequestInfo(r)
	instances := o.Managers()
	for _, instance := range instances {
		fmt.Fprintf(w, "%s\n", instance.PrivateIPAddress)
	}
	instances = o.Workers()
	for _, instance := range instances {
		fmt.Fprintf(w, "%s\n", instance.PrivateIPAddress)
	}
}

// DtrInfo prints the list of Instance ID and their private IP
func (o OracleWeb) DtrInfo(w http.ResponseWriter, r *http.Request) {
	// show DTR image used on manager
	fmt.Fprintf(w, "%s\n", getDTRImage())
}

func oracleInstances(coreClient *core.Client, customFilters core.InstancesParameters) []WebInstance {
	// get the instances from Oracle, takes a filter to limit the results.
	var instances []WebInstance

	result, err := coreClient.ListInstances(&customFilters)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		logrus.Error(err)
		return instances
	}
	logrus.Infof("Instances: %v", result)

	for _, instance := range result {
		logrus.Info("Instance: ", instance.ID)
		vNics := coreClient.ListVNic(instance.ID)
		logrus.Infof("VNics: %v", vNics)

		oInstance := WebInstance{
			InstanceID:       instance.ID,
			InstanceType:     instance.Shape,
			PublicIPAddress:  vNics[0].PublicIP,
			PrivateIPAddress: vNics[0].PrivateIP,
			InstanceState:    instance.LifeCycleState,
			InstanceAZ:       instance.Region,
		}
		instances = append(instances, oInstance)
	}
	return instances
}

// NewOracleClient returns a struct of the web struct
func NewOracleClient() OracleWeb {
	// Get ENV VARS
	env := map[string]string{
		"USER_OCID":      os.Getenv("USER_OCID"),
		"COMPONENT_OCID": os.Getenv("COMPONENT_OCID"),
		"TENANT_OCID":    os.Getenv("TENANT_OCID"),
		"FINGERPRINT":    os.Getenv("FINGERPRINT"),
		"KEY_FILE":       os.Getenv("KEY_FILE"),
		"REGION":         os.Getenv("REGION"),
		"STACK_NAME":     os.Getenv("STACK_NAME"),
	}
	config := &bmc.Config{
		User:        bmc.String(env["USER_OCID"]),
		Fingerprint: bmc.String(env["FINGERPRINT"]),
		KeyFile:     bmc.String(env["KEY_FILE"]),
		Tenancy:     bmc.String(env["TENANT_OCID"]),
		// Pass the region or let the instance metadata dictate it
		Region:   bmc.String(env["REGION"]),
		LogLevel: bmc.LogLevel(bmc.LogDebug),
	}
	// Create the API Client
	apiClient, err := api.NewClient(config)
	if err != nil {
		logrus.Errorf("Error creating OPC Core Services Client: %s", err)
	}
	coreClient := core.NewClient(apiClient, env["COMPONENT_OCID"])
	return OracleWeb{
		Client:      coreClient,
		StackName:   env["STACK_NAME"],
		ComponentID: env["COMPONENT_OCID"],
	}
}

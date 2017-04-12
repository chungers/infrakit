package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/examples/helpers"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/go-autorest/autorest/azure"
)

// AzureWeb interface for all Azure calls
type AzureWeb struct {
}

// TokenManager obtain the Azure Swarm Manager Token
func (a AzureWeb) TokenManager(w http.ResponseWriter, r *http.Request) {
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

// TokenWorker obtain the Azure Swarm Worker Token
func (a AzureWeb) TokenWorker(w http.ResponseWriter, r *http.Request) {
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

// Managers get list of Azure manager instances
func (a AzureWeb) Managers() []WebInstance {
	// get the clients for Network and Compute
	env := map[string]string{
		"AZURE_CLIENT_ID":       os.Getenv("APP_ID"),
		"AZURE_CLIENT_SECRET":   os.Getenv("APP_SECRET"),
		"AZURE_SUBSCRIPTION_ID": os.Getenv("ACCOUNT_ID"),
		"AZURE_TENANT_ID":       os.Getenv("TENANT_ID"),
		"AZURE_GROUP_NAME":      os.Getenv("GROUP_NAME"),
		"AZURE_VMSS_MGR":        os.Getenv("VMSS_MGR"),
		"AZURE_VMSS_WRK":        os.Getenv("VMSS_WRK")}
	nicClient, vmssClient := initClients(env)
	// Get list of VMSS Network Interfaces for Managers
	managerIPTable, err := getVMSSNic(nicClient, env, env["AZURE_VMSS_MGR"])
	if err != nil {
		fmt.Printf("Couldn't get Manager Nic for VMSS: %v", err)
		return []WebInstance{}
	}
	// Get list of VMSS for Managers
	managerVMs, err := getVMSSList(vmssClient, env, env["AZURE_VMSS_MGR"], managerIPTable)
	if err != nil {
		fmt.Printf("Couldn't get List of Manager VMSS: %v", err)
		return []WebInstance{}
	}
	return managerVMs
}

// Workers get list of Azure worker instances
func (a AzureWeb) Workers() []WebInstance {
	// get the clients for Network and Compute
	env := map[string]string{
		"AZURE_CLIENT_ID":       os.Getenv("APP_ID"),
		"AZURE_CLIENT_SECRET":   os.Getenv("APP_SECRET"),
		"AZURE_SUBSCRIPTION_ID": os.Getenv("ACCOUNT_ID"),
		"AZURE_TENANT_ID":       os.Getenv("TENANT_ID"),
		"AZURE_GROUP_NAME":      os.Getenv("GROUP_NAME"),
		"AZURE_VMSS_MGR":        os.Getenv("VMSS_MGR"),
		"AZURE_VMSS_WRK":        os.Getenv("VMSS_WRK")}
	nicClient, vmssClient := initClients(env)
	// Get list of VMSS Network Interfaces for Managers
	workerIPTable, err := getVMSSNic(nicClient, env, env["AZURE_VMSS_WRK"])
	if err != nil {
		fmt.Printf("Couldn't get Worker Nic for VMSS: %v", err)
		return []WebInstance{}
	}
	// Get list of VMSS for Managers
	workerVMs, err := getVMSSList(vmssClient, env, env["AZURE_VMSS_WRK"], workerIPTable)
	if err != nil {
		fmt.Printf("Couldn't get List of Worker VMSS: %v", err)
		return []WebInstance{}
	}
	return workerVMs
}

// Instances prints the list of Instance ID and their private IP
func (a AzureWeb) Instances(w http.ResponseWriter, r *http.Request) {
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
func (a AzureWeb) DtrInfo(w http.ResponseWriter, r *http.Request) {
	// show DTR image used on manager
	fmt.Fprintf(w, "%s\n", getDTRImage())
}

func initClients(env map[string]string) (network.InterfacesClient, compute.VirtualMachineScaleSetVMsClient) {

	spt, err := helpers.NewServicePrincipalTokenFromCredentials(env, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		fmt.Printf("ERROR: Getting SP token - check that all ENV variables are set")
		os.Exit(1)
	}
	// Create Network Interface Client
	nicClient := network.NewInterfacesClient(env["AZURE_SUBSCRIPTION_ID"])
	nicClient.Authorizer = spt
	// Create VMSS Client
	vmssClient := compute.NewVirtualMachineScaleSetVMsClient(env["AZURE_SUBSCRIPTION_ID"])
	vmssClient.Authorizer = spt
	return nicClient, vmssClient
}

func getVMSSNic(client network.InterfacesClient, env map[string]string, vmss string) (IPTable map[string]string, err error) {
	result, err := client.ListVirtualMachineScaleSetNetworkInterfaces(env["AZURE_GROUP_NAME"], vmss)
	if err != nil {
		// Message from an error.
		fmt.Println("Error: ", err.Error())
		return IPTable, err
	}

	IPTable = map[string]string{}

	for _, nic := range *result.Value {
		if *nic.InterfacePropertiesFormat.Primary {
			for _, ipConfig := range *nic.InterfacePropertiesFormat.IPConfigurations {
				if *ipConfig.InterfaceIPConfigurationPropertiesFormat.Primary {
					IPTable[*nic.ID] = *ipConfig.InterfaceIPConfigurationPropertiesFormat.PrivateIPAddress
				}
			}
		}
	}
	return IPTable, nil
}

func getVMSSList(client compute.VirtualMachineScaleSetVMsClient, env map[string]string, vmss string, nicIPTable map[string]string) ([]WebInstance, error) {
	vms := []WebInstance{}

	result, err := client.List(env["AZURE_GROUP_NAME"], vmss, "", "", "")
	if err != nil {
		// Message from an error.
		fmt.Println("Error: ", err.Error())
		return vms, err
	}

	for _, vm := range *result.Value {
		nics := *vm.VirtualMachineScaleSetVMProperties.NetworkProfile.NetworkInterfaces
		privateIP := nicIPTable[*nics[0].ID]
		newVM := WebInstance{
			ID:               *vm.ID,
			InstanceID:       *vm.InstanceID,
			InstanceName:     *vm.Name,
			InstanceType:     *vm.Type,
			InstanceNic:      *nics[0].ID,
			PrivateIPAddress: privateIP,
			InstanceState:    *vm.VirtualMachineScaleSetVMProperties.ProvisioningState}
		vms = append(vms, newVM)
	}
	return vms, nil
}

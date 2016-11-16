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
		"AZURE_SUBSCRIPTION_ID": os.Getenv("SUBSCRIPTION_ID"),
		"AZURE_TENANT_ID":       os.Getenv("TENANT_ID"),
		"AZURE_GROUP_NAME":      os.Getenv("GROUP_NAME"),
		"AZURE_PREFIX":          os.Getenv("PREFIX")}
	nicClient, vmssClient := initClients(env)
	// Get list of VMSS Network Interfaces for Managers
	managerIPTable, err := getVMSSNic(nicClient, env, "managervmss")
	if err != nil {
		fmt.Printf("Couldn't get Manager Nic for VMSS: %v", err)
		return []WebInstance{}
	}
	// Get list of VMSS for Managers
	managerVMs, err := getVMSSList(vmssClient, env, "managervmss", managerIPTable)
	if err != nil {
		fmt.Printf("Couldn't get List of Manager VMSS: %v", err)
	}
	return managerVMs
}

// Workers get list of Azure worker instances
func (a AzureWeb) Workers() []WebInstance {
	// get the clients for Network and Compute
	env := map[string]string{
		"AZURE_CLIENT_ID":       os.Getenv("APP_ID"),
		"AZURE_CLIENT_SECRET":   os.Getenv("APP_SECRET"),
		"AZURE_SUBSCRIPTION_ID": os.Getenv("SUBSCRIPTION_ID"),
		"AZURE_TENANT_ID":       os.Getenv("TENANT_ID"),
		"AZURE_GROUP_NAME":      os.Getenv("GROUP_NAME"),
		"AZURE_PREFIX":          os.Getenv("PREFIX")}
	nicClient, vmssClient := initClients(env)
	// Get list of VMSS Network Interfaces for Managers
	workerIPTable, err := getVMSSNic(nicClient, env, "worker-vmss")
	if err != nil {
		fmt.Errorf("Couldn't get Worker Nic for VMSS: %v", err)
		return []WebInstance{}
	}
	// Get list of VMSS for Managers
	workerVMs, err := getVMSSList(vmssClient, env, "worker-vmss", workerIPTable)
	if err != nil {
		fmt.Printf("Couldn't get List of Worker VMSS: %v", err)
	}
	return workerVMs
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

func getVMSSNic(client network.InterfacesClient, env map[string]string, vmType string) (IPTable map[string]string, err error) {
	vmss := fmt.Sprintf("%s-%s", env["AZURE_PREFIX"], vmType)
	result, err := client.ListVirtualMachineScaleSetNetworkInterfaces(env["AZURE_GROUP_NAME"], vmss)
	if err != nil {
		// Message from an error.
		fmt.Println("Error: ", err.Error())
		return IPTable, err
	}

	IPTable = map[string]string{}

	for _, nic := range *result.Value {
		if *nic.Properties.Primary {
			for _, ipConfig := range *nic.Properties.IPConfigurations {
				if *ipConfig.Properties.Primary {
					fmt.Printf("Adding: %s to table at index: %s\n\n", *ipConfig.Properties.PrivateIPAddress, *nic.ID)
					IPTable[*nic.ID] = *ipConfig.Properties.PrivateIPAddress
				}
			}
		}
	}
	return IPTable, nil
}

func getVMSSList(client compute.VirtualMachineScaleSetVMsClient, env map[string]string, vmType string, nicIPTable map[string]string) ([]WebInstance, error) {
	vmss := fmt.Sprintf("%s-%s", env["AZURE_PREFIX"], vmType)
	vms := []WebInstance{}

	result, err := client.List(env["AZURE_GROUP_NAME"], vmss, "", "", "")
	if err != nil {
		// Message from an error.
		fmt.Println("Error: ", err.Error())
		return vms, err
	}

	for _, vm := range *result.Value {
		nics := *vm.Properties.NetworkProfile.NetworkInterfaces
		privateIP := nicIPTable[*nics[0].ID]
		newVM := WebInstance{
			ID:               *vm.ID,
			InstanceID:       *vm.InstanceID,
			InstanceName:     *vm.Name,
			InstanceType:     *vm.Type,
			InstanceNic:      *nics[0].ID,
			PrivateIPAddress: privateIP,
			InstanceState:    *vm.Properties.ProvisioningState}
		vms = append(vms, newVM)
	}
	return vms, nil
}

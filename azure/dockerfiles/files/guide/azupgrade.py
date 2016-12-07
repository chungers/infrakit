#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
import pytz
from datetime import datetime
from time import sleep
from docker import Client
from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.storage import StorageManagementClient
from azure.mgmt.compute import ComputeManagementClient
from azure.mgmt.network import NetworkManagementClient
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.resource.resources.models import DeploymentMode
from azure.mgmt.storage.models import StorageAccountCreateParameters
from azure.storage.table import TableService, Entity
from azure.storage.queue import QueueService
from azutils import *

APP_SECRET_PARAMETER_NAME_IN_TEMPLATE = 'adServicePrincipalAppSecret'

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']

RG_NAME = os.environ['GROUP_NAME']
SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']

VMSS_ROLE_MAPPING = {
    MGR_VMSS_NAME: 'manager',
    WRK_VMSS_NAME: 'worker'
}

def update_deployment_template(template_url, resource_client):

    print("Updating Resource Group: {}".format(RG_NAME))

    deployments = resource_client.deployments.list(RG_NAME)
    latest_timestamp = datetime.min
    latest_deployment = None

    deployments = resource_client.deployments.list(RG_NAME)
    latest_timestamp = datetime.min.replace(tzinfo=pytz.UTC)
    latest_deployment = None

    for deployment in deployments:
        state = deployment.properties.provisioning_state
        print(("Inspecting deployment: {} at state {}").format(
                deployment.name, state))
        if state != "Succeeded":
            continue
        timestamp = deployment.properties.timestamp

        if timestamp > latest_timestamp:
            latest_timestamp = timestamp
            latest_deployment = deployment

    if latest_deployment is None:
        print("No successful deployment found for {}".format(RG_NAME))
        return

    print("Found deployment: {} deployed at {}".format(
            latest_deployment.name, latest_deployment.properties.timestamp))

    # update link to latest template.json
    template_uri = {
        'uri': template_url
    }

    deployment_properties = {
        'mode': DeploymentMode.incremental,
        'template_link': template_uri
    }

    if latest_deployment.properties.parameters is None:
        print("ERROR: Upgrading Deployment with Parameter Link Not Supported")
        print("Parameters need to be specified as part of the deployment")
        return

    deployment_properties['parameters'] = {}
    print("Using Parameters Object from previous deployment {}".format(
            latest_deployment.properties.parameters))
    # azure requires the type field to be removed from the JSON
    # certain fields like secure strings won't have a value
    for k, v in latest_deployment.properties.parameters.items():
        if 'value' in v:
            print("Adding parameter: {} value: {}".format(k, v['value']))
            deployment_properties['parameters'][k] = {'value': v['value']}
        else:
            print("Skipping parameter: {} contents: {}".format(k, v))

    # secure strings need to be populated separately as they are not returned
    # APP_SECRET is the only secure string parameter user populates now
    deployment_properties['parameters'][APP_SECRET_PARAMETER_NAME_IN_TEMPLATE] = \
        {'value': APP_SECRET}

    async_update = resource_client.deployments.create_or_update(
                        RG_NAME, latest_deployment.name, deployment_properties)
    wait_with_status(async_update, "Waiting for deployment to update ...")

    print("Finished updating deployment: {}".format(latest_deployment.name))


def update_vmss(vmss_name, docker_client, compute_client, network_client, tbl_svc):

    nic_id_table = {}
    vm_ip_table = {}
    vm_id_table = {}

    # Map Azure VMSS instance IDs to swarm node IDs using NIC and IP address
    nics = network_client.network_interfaces.list_virtual_machine_scale_set_network_interfaces(
                                                            RG_NAME, vmss_name)
    for nic in nics:
        # print ("NIC: {} Primary:{}").format(nic.id, nic.primary)
        if nic.primary:
            for ips in nic.ip_configurations:
                # print ("IP: {} Primary:{}").format(ips.private_ip_address, ips.primary)
                if ips.primary:
                    nic_id_table[nic.id] = ips.private_ip_address

    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, vmss_name)
    for vm in vms:
        print("Getting IP of VM: {} in VMSS {}".format(vm.instance_id, vmss_name))
        for nic in vm.network_profile.network_interfaces:
            if nic.id in nic_id_table:
                # print ("IP Address: {}").format(nic_ip_table[nic.id])
                vm_ip_table[nic_id_table[nic.id]] = vm.instance_id

    nodes = docker_client.nodes(filters={'role': VMSS_ROLE_MAPPING[vmss_name]})
    for node in nodes:
        node_ip = node['Status']['Addr']
        print("Node ID: {} IP: {}".format(node['ID'], node_ip))
        if node_ip not in vm_ip_table:
            print("ERROR: Node IP {} not found in list of VM IPs {}".format(
                    node_ip, vm_ip_table))
            return
        vm_id_table[vm_ip_table[node_ip]] = node['ID']

    # at this point all sanity checks and metadata gathering should be complete
    # beyond this point it's hard to recover from any missing data/errors
    for vm in vms:
        node_id = vm_id_table[vm.instance_id]
        print("Accessing VM: {} in VMSS {}".format(vm.instance_id, vmss_name))

        # skip the node we are invoking from for now
        if docker_client.info()['Swarm']['NodeID'] == node_id:
            print("Skip upgrading of manager node where upgrade is initiated from for now")
            continue

        # demote managers and keep track of new leader
        node_info = docker_client.inspect_node(node_id)
        node_hostname = node_info['Description']['Hostname']
        if node_info['Spec']['Role'] == 'manager':
            try:
                leader = node_info['ManagerStatus']['Leader']
            except KeyError:
                leader = False

            subprocess.check_output(["docker", "node", "demote", node_id])
            sleep(5)

            # If node was a leader, update IP of new leader
            # (after demotion of previous leader) in table
            if leader:
                print("Previous Leader demoted. Update leader IP address")
                leader_ip = get_swarm_leader_ip(docker_client)
                update_leader_tbl(tbl_svc, SWARM_TABLE, LEADER_PARTITION,
                                        LEADER_ROW, leader_ip)

        subprocess.check_output(["docker", "node", "rm", "--force", node_id])

        print("Update OS info started for VMSS node: {}".format(vm.instance_id))
        async_vmss_update = compute_client.virtual_machine_scale_sets.update_instances(
                                            RG_NAME, vmss_name, vm.instance_id)
        wait_with_status(async_vmss_update, "Waiting for VM OS info update to complete ...")
        print("Update OS info completed for VMSS node: {}".format(vm.instance_id))

        print "Reimage started for VMSS node: {}".format(vm.instance_id)
        async_vmss_update = compute_client.virtual_machine_scale_set_vms.reimage(
                                            RG_NAME, vmss_name, vm.instance_id)
        wait_with_status(async_vmss_update, "Waiting for VM reimage to complete ...")
        print("Reimage completed for VMSS node: {}".format(vm.instance_id))

        node_booting = True
        while node_booting:
            sleep(10)
            print("Waiting for VMSS node:{} to boot and join back in swarm".format(
                    vm.instance_id))
            for node in docker_client.nodes():
                try:
                    if node['Description']['Hostname'] == node_hostname:
                        node_booting = False
                        break
                except KeyError:
                    # When a member is joining, sometimes things
                    # are a bit unstable and keys are missing. So retry.
                    print("Description/Hostname not found. Retrying ..")
                    continue
        print("VMSS node:{} successfully connected back to swarm".format(
                vm.instance_id))


def main():

    parser = argparse.ArgumentParser(description='Upgrade a Docker4Azure resource group')
    parser.add_argument('template_url', help='New template to upgrade to')

    args = parser.parse_args()

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID
    )

    docker_client = Client(base_url='unix://var/run/docker.sock', version="1.25")

    resource_client = ResourceManagementClient(cred, SUB_ID)
    storage_client = StorageManagementClient(cred, SUB_ID)
    compute_client = ComputeManagementClient(cred, SUB_ID)
    # the default API version for the REST APIs for Network points to 2016-06-01
    # which does not have several VMSS NIC APIs we need. So specify version here
    network_client = NetworkManagementClient(cred, SUB_ID, api_version='2016-09-01')

    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}
    tbl_svc = TableService(account_name=SA_NAME, account_key=storage_keys['key1'])

    qsvc = QueueService(account_name=SA_NAME, account_key=storage_keys['key1'])
    # the Upgrade Msg Queue should only exist when an upgrade is in progress
    if qsvc.exists(UPGRADE_MSG_QUEUE):
        print "Upgrade message queue already exists. Please make sure another upgrade is not in progress."
        return

    qsvc.create_queue(UPGRADE_MSG_QUEUE)

    # Update the resource group template
    print("Updating Resource Group template. This will take several minutes. You can follow the status of the upgrade below or from the Azure console using the URL below:")
    print("https://portal.azure.com/#resource/subscriptions/{}/resourceGroups/{}/overview".format(
            SUB_ID, RG_NAME))
    update_deployment_template(args.template_url, resource_client)

    # Update manager nodes (except the one this script is initiated from)
    print("Starting rolling upgrade of swarm manager nodes. This will take several minutes. You can follow the status of the upgrade below or from the Azure console using the URL below:")
    print("https://portal.azure.com/#resource/subscriptions/{}/resourceGroups/{}/providers/Microsoft.Compute/virtualMachineScaleSets/{}/overview".format(
            SUB_ID, RG_NAME, MGR_VMSS_NAME))
    update_vmss(MGR_VMSS_NAME, docker_client, compute_client, network_client, tbl_svc)

    # Update worker nodes
    print("Starting rolling upgrade of swarm worker nodes. This will take several minutes. You can follow the status of the upgrade below or from the Azure console using the URL below:")
    print("https://portal.azure.com/#resource/subscriptions/{}/resourceGroups/{}/providers/Microsoft.Compute/virtualMachineScaleSets/{}/overview".format(
            SUB_ID, RG_NAME, WRK_VMSS_NAME))
    update_vmss(WRK_VMSS_NAME, docker_client, compute_client, network_client, tbl_svc)

    print("The current VM will be rebooted soon for an upgrade. You can follow the status of the upgrade from the Azure console using the URL below:")
    print("https://portal.azure.com/#resource/subscriptions/{}/resourceGroups/{}/providers/Microsoft.Compute/virtualMachineScaleSets/{}/overview".format(
            SUB_ID, RG_NAME, MGR_VMSS_NAME))
    # Signal another node to upgrade the current node and update leader if necessary
    qsvc.put_message(UPGRADE_MSG_QUEUE, docker_client.info()['Swarm']['NodeID'])

if __name__ == "__main__":
    main()

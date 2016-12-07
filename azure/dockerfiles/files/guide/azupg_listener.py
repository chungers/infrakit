#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
from time import sleep

from docker import Client

from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.storage import StorageManagementClient
from azure.mgmt.compute import ComputeManagementClient
from azure.mgmt.network import NetworkManagementClient
from azure.mgmt.storage.models import StorageAccountCreateParameters
from azure.storage.table import TableService, Entity
from azure.storage.queue import QueueService
from azutils import *

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']

RG_NAME = os.environ['GROUP_NAME']
SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']

def upgrade_mgr_node(node_id, docker_client, compute_client, network_client, storage_key, tbl_svc):

    vmss = compute_client.virtual_machine_scale_sets.get(RG_NAME, MGR_VMSS_NAME)

    node_info = docker_client.inspect_node(node_id)
    node_hostname = node_info['Description']['Hostname']
    try:
        leader = node_info['ManagerStatus']['Leader']
    except KeyError:
        leader = False

    nic_id_table = {}
    vm_ip_table = {}
    node_id_table = {}

    # Find the Azure VMSS instance ID corresponding to the Node ID
    nics = network_client.network_interfaces.list_virtual_machine_scale_set_network_interfaces(
                                                            RG_NAME, MGR_VMSS_NAME)
    for nic in nics:
        # print ("NIC: {} Primary:{}").format(nic.id, nic.primary)
        if nic.primary:
            for ips in nic.ip_configurations:
                # print ("IP: {} Primary:{}").format(ips.private_ip_address, ips.primary)
                if ips.primary:
                    nic_id_table[nic.id] = ips.private_ip_address

    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, MGR_VMSS_NAME)
    for vm in vms:
        print("Getting IP of VM: {} in VMSS {}".format(vm.instance_id, MGR_VMSS_NAME))
        for nic in vm.network_profile.network_interfaces:
            if nic.id in nic_id_table:
                # print ("IP Address: {}").format(nic_ip_table[nic.id])
                vm_ip_table[nic_id_table[nic.id]] = vm.instance_id

    instance_id = -1
    nodes = docker_client.nodes(filters={'role': 'manager'})
    for node in nodes:
        node_ip = node['Status']['Addr']
        print("Node ID: {} IP: {}".format(node['ID'], node_ip))
        if node_ip not in vm_ip_table:
            print("Error: Node IP {} not found in list of VM IPs {}".format(
                    node_ip, vm_ip_table))
            return
        if node['ID'] == node_id:
            instance_id = vm_ip_table[node_ip]

    if instance_id < 0:
        print("ERROR: Node ID:{} could not be mapped to a VMSS Instance ID".format(
                node_id))
        return

    node_info = docker_client.inspect_node(node_id)
    if node_info['Spec']['Role'] == 'manager':
        try:
            leader = node_info['ManagerStatus']['Leader']
        except KeyError:
            leader = False

    # demote the manager node
    subprocess.check_output(["docker", "node", "demote", node_id])
    sleep(5)

    # if leader, update ip
    if leader:
        print("Previous Leader demoted. Update leader IP address")
        leader_ip = get_swarm_leader_ip(docker_client)
        update_leader_tbl(tbl_svc, SWARM_TABLE, LEADER_PARTITION,
                            LEADER_ROW, leader_ip)

    subprocess.check_output(["docker", "node", "rm", "--force", node_id])

    print("Initiating update for instance:{} node:{}".format(instance_id, node_id))
    async_vmss_update = compute_client.virtual_machine_scale_sets.update_instances(
                                            RG_NAME, MGR_VMSS_NAME, instance_id)
    wait_with_status(async_vmss_update, "Waiting for VM OS info update to complete ...")
    print("Update OS info completed for VMSS node: {}".format(instance_id))

    print("Reimage started for VMSS node: {}".format(instance_id))
    async_vmss_update = compute_client.virtual_machine_scale_set_vms.reimage(
                                            RG_NAME, MGR_VMSS_NAME, instance_id)
    wait_with_status(async_vmss_update, "Waiting for VM reimage to complete ...")
    print("Reimage completed for VMSS node: {}".format(instance_id))

    node_booting = True
    while node_booting:
        sleep(10)
        print("Waiting for VMSS node:{} to boot and join back in swarm".format(
                instance_id))
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
    print ("VMSS node:{} successfully connected back to swarm").format(instance_id)


def main():
    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID
    )

    docker_client = Client(base_url='unix://var/run/docker.sock', version="1.25")
    storage_client = StorageManagementClient(cred, SUB_ID)
    compute_client = ComputeManagementClient(cred, SUB_ID)
    # the default API version for the REST APIs for Network points to 2016-06-01
    # which does not have several VMSS NIC APIs we need. So specify version here
    network_client = NetworkManagementClient(cred, SUB_ID, api_version='2016-09-01')

    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}

    tbl_svc = TableService(account_name=SA_NAME, account_key=storage_keys['key1'])
    qsvc = QueueService(account_name=SA_NAME, account_key=storage_keys['key1'])

    if not qsvc.exists(UPGRADE_MSG_QUEUE):
        print("Upgrade message queue not present. Exiting ...")
        return

    msgs = qsvc.peek_messages(UPGRADE_MSG_QUEUE)
    for msg in msgs:
        node_id = msg.content
        print("Obtained Node: {}".format(node_id))
        if docker_client.info()['Swarm']['NodeID'] == node_id:
            print("Recvd msg on the same node we want to upgrade. Skip ..")
            return

    msgs = qsvc.get_messages(UPGRADE_MSG_QUEUE)
    delete_queue = False
    for msg in msgs:
        node_id = msg.content
        print("Obtained Node: {}".format(node_id))
        qsvc.delete_message(UPGRADE_MSG_QUEUE, msg.id, msg.pop_receipt)
        upgrade_mgr_node(node_id, docker_client, compute_client, network_client,
                        storage_keys['key1'], tbl_svc)
        # after successful upgrade, we delete the queue for a fresh new upgrade
        delete_queue = True

    if delete_queue:
        qsvc.delete_queue(UPGRADE_MSG_QUEUE)

if __name__ == "__main__":
    main()

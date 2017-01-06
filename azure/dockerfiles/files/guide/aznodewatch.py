#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
import pytz
import urllib2
import socket
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

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']
ROLE = os.environ['ROLE']

SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']
RG_NAME = os.environ['GROUP_NAME']
IP_ADDR = os.environ['PRIVATE_IP']

RETRY_COUNT = 3
RETRY_INTERVAL = 30

DIAGNOSTICS_MAGIC_PORT = 44554

SWARM_NODE_STATUS_READY = u"ready"

VMSS_ROLE_MAPPING = {
    MGR_VMSS_NAME: 'manager',
    WRK_VMSS_NAME: 'worker'
}


def create_new_vmss_nodes(compute_client, vmss_name, node_count):
    vms_requiring_update = []
    vmss = compute_client.virtual_machine_scale_sets.get(RG_NAME, vmss_name)
    vmss.sku.capacity += node_count
    async_update = compute_client.virtual_machine_scale_sets.create_or_update(
                    RG_NAME, vmss_name, vmss)
    wait_with_status(async_update, "Waiting for vmss capacity to update ...")

    # after the above steps, vmss model gets updated leading to pre-existing
    # VMs reporting not running latest model in the VMSS blade. So update them.
    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, vmss_name)
    for vm in vms:
        if not vm.latest_model_applied:
            print("Add {} to list of VMs to update".format(vm.instance_id))
            vms_requiring_update.append(vm.instance_id)

    async_update = compute_client.virtual_machine_scale_sets.update_instances(
                            RG_NAME, vmss_name, vms_requiring_update)
    wait_with_status(async_update, "Waiting for vmss update to complete ...")


def delete_vmss_node(compute_client, node_id, vmss_name):
    async_update = compute_client.virtual_machine_scale_set_vms.delete(
                    RG_NAME, vmss_name, node_id)
    wait_with_status(async_update, "Waiting for vmss to deallocate vm ...")


def delete_swarm_node(docker_client, ip_addr, role):
    nodes = docker_client.nodes(filters={'role': role})
    for node in nodes:
        try:
            node_ip = node['Status']['Addr']
            node_status = node['Status']['State']
        except:
            # ignore key errors due to phantom/malformed node entries
            continue
        if node_ip == ip_addr:
            node_id = node['ID']
            print("Remove swarm node ID: {} IP: {} Status:{}".format(
                    node_id, node_ip, node_status))
            if node['Spec']['Role'] == 'manager':
                subprocess.check_output(["docker", "node", "demote", node_id])
                sleep(5)
            subprocess.check_output(["docker", "node", "rm", "--force", node_id])


def swarm_node_status(docker_client, ip_addr, role):
    nodes = docker_client.nodes(filters={'role': role})
    for node in nodes:
        try:
            node_ip = node['Status']['Addr']
            node_status = node['Status']['State']
        except:
            # ignore key errors due to phantom/malformed node entries
            continue
        if node_ip == ip_addr:
            return node_status
    return ""


def docker_magic_port_up(ip_addr):
    s = socket.socket()
    for i in range (0, RETRY_COUNT):
        try:
            s.connect((ip_addr, DIAGNOSTICS_MAGIC_PORT))
            s.shutdown(2)
            return True
        except socket.error, e:
            print("Could not reach {} due to: {}".format(ip_addr, e))
            sleep(RETRY_INTERVAL)
    return False


def monitor_vmss_nodes(docker_client, compute_client, network_client, vmss_name):
    nic_id_table = {}
    vm_ip_table = {}
    dead_node_count = 0

    # Find out IP addresses of the nodes in the scale set
    nics = network_client.network_interfaces.list_virtual_machine_scale_set_network_interfaces(
                                                            RG_NAME, vmss_name)
    for nic in nics:
        if nic.primary:
            for ips in nic.ip_configurations:
                if ips.primary:
                    nic_id_table[nic.id] = ips.private_ip_address

    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, vmss_name)
    for vm in vms:
        for nic in vm.network_profile.network_interfaces:
            if nic.id in nic_id_table:
                vm_ip_table[nic_id_table[nic.id]] = vm.instance_id

    # A node is considered dead if it's neither responding to magic port
    # nor reporting as ready in swarm status. If these two conditions are
    # not met, the node could be in an intermediate/transitory state such as
    # restarting .. so ignore it for now, let things settle down and re-examine
    # in another later invocation

    for ip_addr in vm_ip_table.keys():
        if (not docker_magic_port_up(ip_addr) and
            swarm_node_status(docker_client, ip_addr,
                VMSS_ROLE_MAPPING[vmss_name]) != SWARM_NODE_STATUS_READY):
            print("Dead node detected with IP {}".format(ip_addr))
            delete_swarm_node(docker_client, ip_addr, VMSS_ROLE_MAPPING[vmss_name])
            delete_vmss_node(compute_client, vm_ip_table[ip_addr], vmss_name)
            dead_node_count += 1

    if dead_node_count > 0:
        print("Replace {} dead nodes in VMSS".format(dead_node_count))
        create_new_vmss_nodes(compute_client, vmss_name, dead_node_count)


def main():

    # Assumes this is running on a single node: e.g. swarm leader. Whoever kicks
    # this off needs to ensure only one instance of this is running. Multiple
    # invocations of this script may lead to extra nodes being created.

    docker_client = Client(base_url='unix://var/run/docker.sock', version="1.25")

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID
    )

    compute_client = ComputeManagementClient(cred, SUB_ID)
    storage_client = StorageManagementClient(cred, SUB_ID)
    resource_client = ResourceManagementClient(cred, SUB_ID)

    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}

    tbl_svc = TableService(account_name=SA_NAME, account_key=storage_keys['key1'])
    qsvc = QueueService(account_name=SA_NAME, account_key=storage_keys['key1'])

    # don't try to restart nodes if an upgrade is in progress. Nodes are expected
    # to go down during a rolling upgrade and be unresponsive for periods of time
    if qsvc.exists(UPGRADE_MSG_QUEUE):
        return

    # the default API version for the REST APIs for Network points to 2016-06-01
    # which does not have several VMSS NIC APIs we need. So specify version here
    network_client = NetworkManagementClient(cred, SUB_ID, api_version='2016-09-01')

    monitor_vmss_nodes(docker_client, compute_client, network_client, MGR_VMSS_NAME)
    monitor_vmss_nodes(docker_client, compute_client, network_client, WRK_VMSS_NAME)

if __name__ == "__main__":
    main()

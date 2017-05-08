#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
import pytz
import socket
import logging
import logging.config
from datetime import datetime, timedelta
from time import sleep
from docker import Client
from urllib2 import Request, urlopen, URLError, HTTPError
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
from azendpt import *

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']

SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']
RG_NAME = os.environ['GROUP_NAME']

RESOURCE_MANAGER_ENDPOINT = os.getenv('RESOURCE_MANAGER_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['RESOURCE_MANAGER_ENDPOINT'])
ACTIVE_DIRECTORY_ENDPOINT = os.getenv('ACTIVE_DIRECTORY_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['ACTIVE_DIRECTORY_ENDPOINT'])
STORAGE_ENDPOINT = os.getenv('STORAGE_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['STORAGE_ENDPOINT'])

RETRY_COUNT = 3
RETRY_INTERVAL = 30

DIAGNOSTICS_MAGIC_PORT = 44554

SWARM_NODE_STATUS_READY = u"ready"
HTTP_200_OK = 200

VMSS_ROLE_MAPPING = {
    MGR_VMSS_NAME: 'manager',
    WRK_VMSS_NAME: 'worker'
}

AZURE_RUNNING = "PowerState/running"
AZURE_PROVISIONED = "ProvisioningState/succeeded"
# wait 10 mins after provisioning state change before starting health checks
POST_PROVISIONING_WAIT = 600

LOG_CFG_FILE = "/etc/aznodewatch_log_cfg.json"
LOG = logging.getLogger("aznodewatch")

def create_new_vmss_nodes(compute_client, vmss_name, node_count):
    vms_requiring_update = []
    vmss = compute_client.virtual_machine_scale_sets.get(RG_NAME, vmss_name)
    vmss.sku.capacity += node_count
    async_update = compute_client.virtual_machine_scale_sets.create_or_update(
                    RG_NAME, vmss_name, vmss)
    wait_with_status(async_update, u"Waiting for vmss capacity to update ...")

    # after the above steps, vmss model gets updated leading to pre-existing
    # VMs reporting not running latest model in the VMSS blade. So update them.
    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, vmss_name)
    for vm in vms:
        if not vm.latest_model_applied:
            LOG.info(u"Add {} to list of VMs to update".format(vm.instance_id))
            vms_requiring_update.append(vm.instance_id)

    async_update = compute_client.virtual_machine_scale_sets.update_instances(
                            RG_NAME, vmss_name, vms_requiring_update)
    wait_with_status(async_update, u"Waiting for vmss update to complete ...")
    LOG.info(u"VM model update completed")

def delete_vmss_node(compute_client, node_id, vmss_name):
    async_update = compute_client.virtual_machine_scale_set_vms.delete(
                    RG_NAME, vmss_name, node_id)
    wait_with_status(async_update, u"Waiting for vmss to deallocate vm ...")


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
            LOG.info(u"Remove swarm node ID: {} IP: {} Status:{}".format(
                    node_id, node_ip, node_status))
            if node['Spec']['Role'] == 'manager':
                subprocess.check_output(["docker", "node", "demote", node_id])
                sleep(5)
            subprocess.check_output(["docker", "node", "rm", "--force", node_id])


def swarm_node_status(docker_client, ip_addr, role):
    nodes = docker_client.nodes(filters={'role': role})
    for node in nodes:
        node_ip = "0.0.0.0"
        node_status = ""
        try:
            node_ip = node['Status']['Addr']
            node_status = node['Status']['State']
            LOG.info(u"node ip {} swarm status {}".format(node_ip, node_status))
        except:
            # ignore key errors due to phantom/malformed node entries
            continue
        if node_ip == ip_addr:
            return node_status
    return ""


def docker_diagnostics_response(ip_addr):
    req = Request("http://{}:{}".format(ip_addr, DIAGNOSTICS_MAGIC_PORT))
    for i in range (0, RETRY_COUNT):
        try:
            response = urlopen(req)
            return response.getcode()
        except HTTPError as e:
            LOG.warning(u"Could not reach {} due to: {}".format(ip_addr, e.code))
            sleep(RETRY_INTERVAL)
        except URLError as e:
            LOG.warning(u"Could not reach {} due to: {}".format(ip_addr, e.reason))
            sleep(RETRY_INTERVAL)
    return -1


def monitor_vmss_nodes(docker_client, compute_client, network_client, vmss_name):
    nic_id_table = {}
    vm_ip_table = {}
    vm_perform_health_check = {}
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
        provisioned_delta = timedelta.resolution
        vm_running = False
        vm_perform_health_check[vm.instance_id] = False

        # find out how long back VM was provisioned to deal with delays
        # during initial start or reboot
        vm_view = compute_client.virtual_machine_scale_set_vms.get_instance_view(
                                        RG_NAME, vmss_name, vm.instance_id)
        for status in vm_view.statuses:
            # Azure VMs have a power state [running/stopping/deallocating/etc]
            # and a provisioning state with timestamp. Check both.

            # We skip VMs in non running states since those have been subject to
            # explicit action by admin through Azure mgmt and we honor that

            if status.code == AZURE_RUNNING:
                vm_running = True
            if status.code == AZURE_PROVISIONED and status.time != None:
                # timestamp gets reset during reboots (besides initial creation)
                # allowing us to wait longer for things to stabilize after reboot
                provisioned_delta = datetime.utcnow().replace(tzinfo=pytz.utc) - status.time

        if vm_running and provisioned_delta.total_seconds() > POST_PROVISIONING_WAIT:
            vm_perform_health_check[vm.instance_id] = True

        for nic in vm.network_profile.network_interfaces:
            if nic.id in nic_id_table:
                vm_ip_table[nic_id_table[nic.id]] = vm.instance_id


    # A running/provisioned node is considered dead if it's neither responding to
    # magic port nor reporting as ready in swarm status. If these two conditions are
    # not met, the node could be in an intermediate/transitory state such as
    # restarting .. so ignore it for now, let things settle down and re-examine
    # in another later invocation

    for ip_addr in vm_ip_table.keys():
        if (vm_perform_health_check[vm_ip_table[ip_addr]] == True and
            docker_diagnostics_response(ip_addr) != HTTP_200_OK and
            swarm_node_status(docker_client, ip_addr,
                VMSS_ROLE_MAPPING[vmss_name]) != SWARM_NODE_STATUS_READY):
            LOG.info(u"Dead node detected with IP {}".format(ip_addr))
            delete_swarm_node(docker_client, ip_addr, VMSS_ROLE_MAPPING[vmss_name])
            delete_vmss_node(compute_client, vm_ip_table[ip_addr], vmss_name)
            dead_node_count += 1

    if dead_node_count > 0:
        LOG.info(u"Replace {} dead nodes in VMSS".format(dead_node_count))
        create_new_vmss_nodes(compute_client, vmss_name, dead_node_count)


def cleanup_swarm_nodes(docker_client, compute_client, network_client, vmss_name):
    ip_list = []

    # Find out IP addresses of NICs in the scale set
    nics = network_client.network_interfaces.list_virtual_machine_scale_set_network_interfaces(
                                                            RG_NAME, vmss_name)
    for nic in nics:
        if nic.primary:
            for ips in nic.ip_configurations:
                if ips.primary:
                    ip_list.append(ips.private_ip_address)

    nodes = docker_client.nodes(filters={'role': VMSS_ROLE_MAPPING[vmss_name]})
    for node in nodes:
        node_ip = "0.0.0.0"
        node_status = ""
        try:
            node_ip = node['Status']['Addr']
            node_status = node['Status']['State']
        except:
            # ignore key errors due to phantom/malformed node entries
            continue

        if node_ip in ip_list:
            # node ip is present in vmss ip list so no action necessary here
            continue

        if node_status != SWARM_NODE_STATUS_READY:
            node_id = node['ID']
            LOG.info(u"Remove non existent swarm node ID: {} IP: {} Status:{}".format(
                node_id, node_ip, node_status))
            if node['Spec']['Role'] == 'manager':
                LOG.info(u"Demote manager first")
                subprocess.check_output(["docker", "node", "demote", node_id])
                sleep(5)
            LOG.info(u"Remove node from swarm")
            subprocess.check_output(["docker", "node", "rm", "--force", node_id])

def main():

    # Assumes this is running on a single node: e.g. swarm leader. Whoever kicks
    # this off needs to ensure only one instance of this is running. Multiple
    # invocations of this script may lead to extra nodes being created.

    with open(LOG_CFG_FILE) as log_cfg_file:
        log_cfg = json.load(log_cfg_file)
        logging.config.dictConfig(log_cfg)

    LOG.debug("Node monitor started")

    docker_client = Client(base_url='unix://var/run/docker.sock', version="1.25")

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID,
        resource=RESOURCE_MANAGER_ENDPOINT,
        auth_uri=ACTIVE_DIRECTORY_ENDPOINT
    )

    compute_client = ComputeManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)
    storage_client = StorageManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)
    resource_client = ResourceManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)

    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}

    tbl_svc = TableService(account_name=SA_NAME, account_key=storage_keys['key1'], endpoint_suffix=STORAGE_ENDPOINT)
    qsvc = QueueService(account_name=SA_NAME, account_key=storage_keys['key1'], endpoint_suffix=STORAGE_ENDPOINT)

    # don't try to restart nodes if an upgrade is in progress. Nodes are expected
    # to go down during a rolling upgrade and be unresponsive for periods of time
    if qsvc.exists(UPGRADE_MSG_QUEUE):
        LOG.info("Upgrade is in progress. Exit")
        return

    # the default API version for the REST APIs for Network points to 2016-06-01
    # which does not have several VMSS NIC APIs we need. So specify version here
    network_client = NetworkManagementClient(cred, SUB_ID, api_version='2016-09-01', base_url=RESOURCE_MANAGER_ENDPOINT)

    # cleanup swarm nodes that do not exist in vmss
    cleanup_swarm_nodes(docker_client, compute_client, network_client, MGR_VMSS_NAME)
    cleanup_swarm_nodes(docker_client, compute_client, network_client, WRK_VMSS_NAME)

    # detect dead nodes and spin up replacements
    monitor_vmss_nodes(docker_client, compute_client, network_client, MGR_VMSS_NAME)
    monitor_vmss_nodes(docker_client, compute_client, network_client, WRK_VMSS_NAME)

if __name__ == "__main__":
    main()

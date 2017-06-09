#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
import logging
import logging.config
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
from dockerutils import *
from azendpt import AZURE_PLATFORMS, AZURE_DEFAULT_ENV

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']
ROLE = os.environ['ROLE']

RG_NAME = os.environ['GROUP_NAME']
SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']

LOG_CFG_FILE = "/etc/azupg_listener_log_cfg.json"
LOG = logging.getLogger("azupg_listener")

RESOURCE_MANAGER_ENDPOINT = os.getenv('RESOURCE_MANAGER_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['RESOURCE_MANAGER_ENDPOINT'])
ACTIVE_DIRECTORY_ENDPOINT = os.getenv('ACTIVE_DIRECTORY_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['ACTIVE_DIRECTORY_ENDPOINT'])
STORAGE_ENDPOINT = os.getenv('STORAGE_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['STORAGE_ENDPOINT'])

def get_manager_count(compute_client):
    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, MGR_VMSS_NAME)
    nodes = 0
    # vms is not a regular list which can be passed to len .. instead it's paged
    for vm in vms:
        nodes += 1
    return nodes

def get_single_manager_instance_id(compute_client):
    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, MGR_VMSS_NAME)
    instance_id = 0
    # expect just one. vms[0] does not work since it's not a regular list. so iterate
    for vm in vms:
        instance_id = vm.instance_id
    return instance_id

def notify_workers_to_rejoin_swarm(compute_client, network_client, qsvc):
    LOG.info("Initiating swarm rejoin. Create queue for notifications")
    qsvc.create_queue(REJOIN_MSG_QUEUE, fail_on_exist=False)

    nic_id_table = {}
    # Find the Azure VMSS instance ID corresponding to the Node ID
    nics = network_client.network_interfaces.list_virtual_machine_scale_set_network_interfaces(
                                                            RG_NAME, WRK_VMSS_NAME)
    for nic in nics:
        if nic.primary:
            for ips in nic.ip_configurations:
                if ips.primary:
                    nic_id_table[nic.id] = ips.private_ip_address

    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, WRK_VMSS_NAME)
    for vm in vms:
        for nic in vm.network_profile.network_interfaces:
            if nic.id in nic_id_table:
                qsvc.put_message(REJOIN_MSG_QUEUE, nic_id_table[nic.id])

    LOG.info("Monitor rejoin queue")
    # let things settle down for a bit in the queue and items be consumed
    sleep(300)
    delete_queue = False
    while not delete_queue:
        sleep(120)
        metadata = qsvc.get_queue_metadata(REJOIN_MSG_QUEUE)
        count = metadata.approximate_message_count
        if count == 0:
            delete_queue = True

    LOG.info("Delete rejoin queue")
    qsvc.delete_queue(REJOIN_MSG_QUEUE)


def upgrade_azure_node(compute_client, instance_id, node_hostname):
    LOG.info("Initiating update for instance:{}".format(instance_id))
    async_vmss_update = compute_client.virtual_machine_scale_sets.update_instances(
                                            RG_NAME, MGR_VMSS_NAME, instance_id)
    wait_with_status(async_vmss_update, "Waiting for VM OS info update to complete ...")
    LOG.info("Update OS info completed for VMSS node: {}".format(instance_id))

    LOG.info("Reimage started for VMSS node: {}".format(instance_id))
    async_vmss_update = compute_client.virtual_machine_scale_set_vms.reimage(
                                            RG_NAME, MGR_VMSS_NAME, instance_id)
    while True:
        if async_vmss_update.done():
            break
        LOG.info("Waiting for VM reimage to complete ...")
        sleep(10)
        # During the reimage phase, Azure spins up multiple nodes even 
        # if overprovision is explicitly set to false on the VMSS
        # Check for these nodes and remove them from swarm so that the 
        # successful node can get the token from metadata server
        remove_overprovisioned_nodes(docker_client, node_hostname, LOG)
    LOG.info("Reimage completed for VMSS node: {}".format(vm.instance_id))
    # Check for overprovisioned nodes blocking the successful node once again
    # in case we did not detect/clean up above.
    remove_overprovisioned_nodes(docker_client, node_hostname, LOG)


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
        LOG.info("NIC ID: {} Primary:{}".format(nic.id, nic.primary))
        if nic.primary:
            for ips in nic.ip_configurations:
                LOG.info("IP: {} Primary:{}".format(ips.private_ip_address, ips.primary))
                if ips.primary:
                    nic_id_table[nic.id] = ips.private_ip_address

    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, MGR_VMSS_NAME)
    for vm in vms:
        LOG.info("Get IP of VM: {} in VMSS {}".format(vm.instance_id, MGR_VMSS_NAME))
        for nic in vm.network_profile.network_interfaces:
            if nic.id in nic_id_table:
                LOG.info("IP Address: {}".format(nic_id_table[nic.id]))
                vm_ip_table[nic_id_table[nic.id]] = vm.instance_id

    instance_id = -1
    nodes = docker_client.nodes(filters={'role': 'manager'})
    for node in nodes:
        node_ip = node['Status']['Addr']
        LOG.info("Node ID: {} IP: {}".format(node['ID'], node_ip))
        if node_ip not in vm_ip_table:
            LOG.error("Node IP {} not found in list of VM IPs {}".format(
                    node_ip, vm_ip_table))
            return
        if node['ID'] == node_id:
            instance_id = vm_ip_table[node_ip]

    if instance_id < 0:
        LOG.error("Node ID:{} could not be mapped to a VMSS Instance ID".format(
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
        LOG.info("Previous Leader demoted. Update leader IP address")
        leader_ip = get_swarm_leader_ip(docker_client)
        update_leader_tbl(tbl_svc, SWARM_TABLE, LEADER_PARTITION,
                            LEADER_ROW, leader_ip)

    subprocess.check_output(["docker", "node", "rm", "--force", node_id])

    # call the core Azure APIs to upgrade the node
    upgrade_azure_node(compute_client, instance_id, node_hostname)

    node_ready = False
    while not node_ready:
        sleep(10)
        LOG.info("Waiting for VMSS node:{} to boot and join back in swarm".format(
                instance_id))
        for node in docker_client.nodes():
            try:
                if (node['Description']['Hostname'] == node_hostname) and (node['Status']['State'] == DOCKER_NODE_STATUS_READY):
                    node_ready = True
                    break
            except KeyError:
                # When a member is joining, sometimes things
                # are a bit unstable and keys are missing. So retry.
                LOG.info("Description/Hostname not found. Retrying ..")
                continue
    LOG.info("VMSS node:{} successfully connected back to swarm".format(instance_id))


def main():

    with open(LOG_CFG_FILE) as log_cfg_file:
        log_cfg = json.load(log_cfg_file)
        logging.config.dictConfig(log_cfg)

    LOG.debug("Upgrade listener started")

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID,
        resource=RESOURCE_MANAGER_ENDPOINT,
        auth_uri=ACTIVE_DIRECTORY_ENDPOINT
    )

    docker_client = Client(base_url='unix://var/run/docker.sock', version="1.25")
    storage_client = StorageManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)
    compute_client = ComputeManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)
    # the default API version for the REST APIs for Network points to 2016-06-01
    # which does not have several VMSS NIC APIs we need. So specify version here
    network_client = NetworkManagementClient(cred, SUB_ID, api_version='2016-09-01', base_url=RESOURCE_MANAGER_ENDPOINT)

    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}

    tbl_svc = TableService(account_name=SA_NAME, account_key=storage_keys['key1'], endpoint_suffix=STORAGE_ENDPOINT)
    qsvc = QueueService(account_name=SA_NAME, account_key=storage_keys['key1'], endpoint_suffix=STORAGE_ENDPOINT)

    if not qsvc.exists(UPGRADE_MSG_QUEUE):
        LOG.debug("Upgrade message queue not present")
        return

    metadata = qsvc.get_queue_metadata(UPGRADE_MSG_QUEUE)
    count = metadata.approximate_message_count
    if count == 0:
        LOG.debug("Nothing detected in upgrade message queue")
        return

    LOG.info("Process messages in upgrade message queue")
    msgs = qsvc.peek_messages(UPGRADE_MSG_QUEUE)
    for msg in msgs:
        node_id = msg.content
        LOG.info("Check node ID. Obtained Node ID: {}".format(node_id))
        if docker_client.info()['Swarm']['NodeID'] == node_id:
            LOG.info("Recvd msg on the same node we want to upgrade. Skip ..")
            return

    if get_manager_count(compute_client) == 1 and ROLE == WRK_ROLE:
        LOG.info("Single manager swarm upgrade scenario detected")
        msgs = qsvc.get_messages(UPGRADE_MSG_QUEUE)
        perform_upgrade = False
        for msg in msgs:
            # no need to look at node id since worker can't do anything with it
            LOG.info("Delete message in upgrade queue and proceed with upgrade")
            qsvc.delete_message(UPGRADE_MSG_QUEUE, msg.id, msg.pop_receipt)
            perform_upgrade = True

        # multiple worker nodes will reach here even if they didn't dequeue msg
        # so set a flag above and only proceed for the node that did dequeue
        if perform_upgrade:
            # delete the swarm table that gets created by leader/manager
            LOG.info("Delete swarm table")
            tbl_svc.delete_table(SWARM_TABLE)
            # directly call the core azure upgrade node since there is a single manager
            LOG.info("Upgrade single leader node")
            upgrade_azure_node(compute_client, get_single_manager_instance_id(compute_client))
            LOG.info("Notify workers to leave and rejoin swarm")
            notify_workers_to_rejoin_swarm(compute_client, network_client, qsvc)
            delete_queue = True
        else:
            LOG.info("Not selected to perform upgrade. Backing off")

    if ROLE == MGR_ROLE:
        LOG.info("Multi manager swarm upgrade scenario detected")
        msgs = qsvc.get_messages(UPGRADE_MSG_QUEUE)
        delete_queue = False
        for msg in msgs:
            node_id = msg.content
            LOG.info("Node ID to upgrade: {}".format(node_id))
            qsvc.delete_message(UPGRADE_MSG_QUEUE, msg.id, msg.pop_receipt)
            LOG.info("Upgrade the manager node upgrade was triggered from")
            upgrade_mgr_node(node_id, docker_client, compute_client, network_client,
                                storage_keys['key1'], tbl_svc)
            # after successful upgrade, we delete the queue for a fresh new upgrade
            delete_queue = True

    if delete_queue:
        LOG.info("Delete the upgrade message queue")
        qsvc.delete_queue(UPGRADE_MSG_QUEUE)

if __name__ == "__main__":
    main()

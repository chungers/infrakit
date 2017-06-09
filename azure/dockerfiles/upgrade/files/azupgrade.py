#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
import pytz
import urllib2
import ssl
import logging
import logging.config
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
from dockerutils import *
from azendpt import AZURE_PLATFORMS, AZURE_DEFAULT_ENV

################################################################
## Expected execution environment: editions_guide container   ##
## The upgrade container packaging this file does not exec it ##
## Dockerfile for editions_guide container needs to have all  ##
## dependencies for azupgrade.py to run properly              ##
################################################################

LOG_CFG_FILE = "/etc/azupgrade_log_cfg.json"
LOG = logging.getLogger("azupg")

APP_SECRET_PARAMETER_NAME_IN_TEMPLATE = 'adServicePrincipalAppSecret'

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']

RG_NAME = os.environ['GROUP_NAME']
SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']

#DDC params
HAS_DDC = False
DTR_TBL_NAME = 'dtrtable'
DTR_PARTITION_NAME = 'dtrreplicas'
LAST_MANAGER_NODE_ID = ''
PRODUCTION_HUB_NAMESPACE = 'docker'
HUB_NAMESPACE = 'docker'
UCP_HUB_TAG = '2.0.2'
DTR_HUB_TAG = '2.1.0'
UCP_IMAGE = '%s/ucp:%s' % (HUB_NAMESPACE,UCP_HUB_TAG)
DTR_IMAGE = '%s/dtr:%s' % (HUB_NAMESPACE,DTR_HUB_TAG)
DTR_PORT = 443
UCP_PORT = 8443

RESOURCE_MANAGER_ENDPOINT = os.getenv('RESOURCE_MANAGER_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['RESOURCE_MANAGER_ENDPOINT'])
ACTIVE_DIRECTORY_ENDPOINT = os.getenv('ACTIVE_DIRECTORY_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['ACTIVE_DIRECTORY_ENDPOINT'])
STORAGE_ENDPOINT = os.getenv('STORAGE_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['STORAGE_ENDPOINT'])

try:
    SA_DTR_NAME = os.environ['DTR_STORAGE_ACCOUNT']
    UCP_ADMIN_USER = os.environ['UCP_ADMIN_USER']
    UCP_ADMIN_PASSWORD = os.environ['UCP_ADMIN_PASSWORD']
    UCP_ELB_HOSTNAME = os.environ['UCP_ELB_HOSTNAME']
    HAS_DDC = True
except:
    pass

VMSS_ROLE_MAPPING = {
    MGR_VMSS_NAME: 'manager',
    WRK_VMSS_NAME: 'worker'
}


# update description on DTR table storage
def set_upgrade_desc(sa_key, desc):
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key, endpoint_suffix=STORAGE_ENDPOINT)
    if not tbl_svc.exists(DTR_TBL_NAME):
        return False
    try:
        replicaids = tbl_svc.query_entities(DTR_TBL_NAME, filter="PartitionKey eq 'dtrreplicas'")
        for replicaid in replicaids:
                upgrade_desc = {'PartitionKey': 'dtrreplicas', 'RowKey': replicaid.replica_id, 'replica_id': replicaid.replica_id, 'node_name': replicaid.node_name, 'description': desc}
                tbl_svc.insert_or_replace_entity(DTR_TBL_NAME, upgrade_desc)
    except:
        LOG.error( "exception while updating replica-id description")
        return False

# get all replica ids from DTR Table
def get_ids(sa_key):
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key, endpoint_suffix=STORAGE_ENDPOINT)
    if not tbl_svc.exists(DTR_TBL_NAME):
        return False
    try:
        replicaids = tbl_svc.query_entities(DTR_TBL_NAME, filter="PartitionKey eq 'dtrreplicas'")
        for replicaid in replicaids:
                LOG.info("{}".format(replicaid.replica_id))
        return replicaids
    except:
        return False

# get replica id that matches swarm node from DTR Table
def get_id(sa_key, nodename):
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key, endpoint_suffix=STORAGE_ENDPOINT)
    if not tbl_svc.exists(DTR_TBL_NAME):
        return False
    try:
        replicaids = tbl_svc.query_entities(DTR_TBL_NAME, filter="PartitionKey eq 'dtrreplicas'")
        if nodename is None:
                return replicaids
        for replicaid in replicaids:
                if replicaid.node_name == nodename:
                        LOG.info("{}".format(replicaid.replica_id))
                        return replicaid.replica_id
    except:
        return False

#delete replicaid from the DTR Table
def delete_id(sa_key, replica_id):
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key, endpoint_suffix=STORAGE_ENDPOINT)
    try:
        # this upsert operation should always succeed
        tbl_svc.delete_entity(DTR_TBL_NAME, DTR_PARTITION_NAME, replica_id)
        LOG.info("successfully deleted replica-id")
        return True
    except:
        LOG.error("exception while deleting replica-id")
        return False

# Checking if UCP is up and running
def checkUCP(client):
    LOG.info("Checking to see if UCP is up and healthy")
    n = 0
    while n < 20:
        LOG.info("Checking managers. Try #{}".format(n))
        nodes = client.nodes(filters={'role': 'manager'})
        # Find first node that's not myself
        ALLGOOD = 'yes'
        for node in nodes:
            Manager_IP = node['Status']['Addr']
            LOG.info("Manager IP: {}".format(Manager_IP))
            # Checking if UCP is up and running
            UCP_URL = 'https://%s:%s/_ping' %(Manager_IP, UCP_PORT)
            LOG.info("{}".format(UCP_URL))
            ssl._create_default_https_context = ssl._create_unverified_context
            try:
                resp = urllib2.urlopen(UCP_URL)
            except urllib2.URLError, e:
                LOG.info("URLError {}".format(str(e.reason)))
                ALLGOOD = 'no'
            except urllib2.HTTPError, e:
                LOG.info("HTTPError {}".format(str(e.code)))
                ALLGOOD = 'no'
            else:
                LOG.info("UCP on %s is healthy" % Manager_IP)
        if  ALLGOOD == 'yes':
            LOG.info("UCP is all healthy, good to move on!")
            break
        else:
            LOG.info("Not all healthy, rest and try again..")
            if n == 20:
                # this will cause the Autoscale group to timeout, and leave this node in the swarm
                # it will eventually be killed once the timeout it his. TODO: Do something about this.
                LOG.error("UCP failed status check after $n tries. Aborting...")
                exit(0)
            sleep(30)
            n = n + 1


# check if DDC is installed, if so, make sure it is in a stable state before we continue.
def checkDDC(client, node_name):

    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID,
        resource=RESOURCE_MANAGER_ENDPOINT,
        auth_uri=ACTIVE_DIRECTORY_ENDPOINT
    )
    storage_client = StorageManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)
    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_DTR_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}
    key = storage_keys['key1']
    LOG.info("HostName: {}".format(node_name))
    LOG.info("UCP is installed, make sure it is ready, before we continue.")
    checkUCP(client)
    LOG.info("Remove DTR ")
    local_dtr_id = get_id(key, node_name)
    if local_dtr_id is None or local_dtr_id == False:
        LOG.info("DTR not installed on this node")
        return
    LOG.info("LOCAL DTR ID: {}".format(local_dtr_id))
    LOG.info("remove from Azure DTR Table")
    delete_id(key, local_dtr_id)
    num_replicas = 0
    replicas = get_ids(key)
    for replica in replicas:
        num_replicas = num_replicas + 1
        existing_replica_id = replica.replica_id
        LOG.info("Existing Replica ID {}".format(existing_replica_id))
    LOG.info("Num of replicas: {}".format(num_replicas))
    last_manager = 0
    if replicas is None or num_replicas == 0:
        existing_replica_id = local_dtr_id
        last_manager = 1
        LOG.info("Last Existing Replica ID: {}".format(existing_replica_id))
    LOG.info("Remove DTR node.")
    # set response to 1, so we guarantee it goes into until loop at least once.
    response = 1
    count = 1
    # try to remove node, keep trying until we have a good removal status 0
    while response != 0:
        LOG.info("removing DTR node... try#{}".format(count))
        if (last_manager != 1) or (existing_replica_id != local_dtr_id):
                cont = client.create_container(
                image=DTR_IMAGE,
                command='remove --debug --ucp-url https://{elb_host} --ucp-username {user}'
                    ' --ucp-password {password} --ucp-insecure-tls --existing-replica-id {existing_replica}'
                    ' --replica-id {local_dtr_id}'.format(
                        elb_host=UCP_ELB_HOSTNAME,
                        user=UCP_ADMIN_USER,
                        password=UCP_ADMIN_PASSWORD,
                        existing_replica=existing_replica_id,
                        local_dtr_id=local_dtr_id
                        )
                )
        else:
                LOG.info("deleting last DTR node")
                cont = client.create_container(
                image=DTR_IMAGE,
                command='remove --debug --force-remove --ucp-url https://{elb_host} --ucp-username {user}'
                    ' --ucp-password {password} --ucp-insecure-tls --existing-replica-id {existing_replica}'
                    ' --replica-id {local_dtr_id}'.format(
                        elb_host=UCP_ELB_HOSTNAME,
                        user=UCP_ADMIN_USER,
                        password=UCP_ADMIN_PASSWORD,
                        existing_replica=existing_replica_id,
                        local_dtr_id=local_dtr_id
                        )
                )
        LOG.info("{}".format(cont))
        client.start(cont)
        response = client.wait(cont)
        LOG.info("DTR Remove Command Response:{}".format(response))
        if response != 0:
            if count == 20:
                LOG.error("Tried to remove node $count times. We are over limit, aborting...")
                exit(1)
            LOG.info("We failed for a reason, lets retry again after a brief delay.")
            sleep(30)
            count = count + 1
        else:
            LOG.info("Node removal complete")

    LOG.info("Final cleanup check..")
    LOG.info("DTR remove is complete.")
    checkUCP(client)
    LOG.info("UCP is good to go, continue.")
    sleep(10)

def update_lastmanager(client):
    global LAST_MANAGER_NODE_ID
    
    if LAST_MANAGER_NODE_ID is not None:
        try:
            LOG.info("update last manager in try LAST MANAGER Node ID: {}".format(LAST_MANAGER_NODE_ID))
            node_info = client.inspect_node(LAST_MANAGER_NODE_ID)
            node_hostname = node_info['Description']['Hostname']
            LOG.info("Last Manager Hostname: {}".format(node_hostname))
            checkDDC(client, node_hostname)
            sleep(5)
        except:
            LOG.error("Failed to remove last DTR node")

def validate_template(template_url):
   response = urllib2.urlopen(template_url)
   template = json.load(response)

def update_deployment_template(template_url, resource_client):

    LOG.info("Updating Resource Group: {}".format(RG_NAME))

    deployments = resource_client.deployments.list(RG_NAME)
    latest_timestamp = datetime.min
    latest_deployment = None

    deployments = resource_client.deployments.list(RG_NAME)
    latest_timestamp = datetime.min.replace(tzinfo=pytz.UTC)
    latest_deployment = None

    for deployment in deployments:
        state = deployment.properties.provisioning_state
        LOG.info(("Inspecting deployment: {} at state {}").format(
                deployment.name, state))
        if state != "Succeeded":
            continue
        timestamp = deployment.properties.timestamp

        if timestamp > latest_timestamp:
            latest_timestamp = timestamp
            latest_deployment = deployment

    if latest_deployment is None:
        LOG.error("No successful deployment found for {}".format(RG_NAME))
        return

    LOG.info("Found deployment: {} deployed at {}".format(
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
        LOG.error("Upgrading Deployment with Parameter Link Not Supported")
        LOG.error("Parameters need to be specified as part of the deployment")
        return

    deployment_properties['parameters'] = {}
    LOG.info("Using Parameters Object from previous deployment {}".format(
            latest_deployment.properties.parameters))
    # azure requires the type field to be removed from the JSON
    # certain fields like secure strings won't have a value
    for k, v in latest_deployment.properties.parameters.items():
        if 'value' in v:
            LOG.info("Adding parameter: {} value: {}".format(k, v['value']))
            deployment_properties['parameters'][k] = {'value': v['value']}
        else:
            LOG.info("Skipping parameter: {} contents: {}".format(k, v))

    # secure strings need to be populated separately as they are not returned
    # APP_SECRET is the only secure string parameter user populates now
    deployment_properties['parameters'][APP_SECRET_PARAMETER_NAME_IN_TEMPLATE] = \
        {'value': APP_SECRET}

    async_update = resource_client.deployments.create_or_update(
                        RG_NAME, latest_deployment.name, deployment_properties)
    wait_with_status(async_update, "Waiting for deployment to update ...")

    LOG.info("Finished updating deployment: {}".format(latest_deployment.name))

def update_vmss(vmss_name, docker_client, compute_client, network_client, tbl_svc):

    global LAST_MANAGER_NODE_ID

    # tmp lookup table for nic.id -> ip address used for populating other tables
    nic_id_table = {}
    # tmp lookup table for ip -> instance_id used for populating other tables
    vm_tmp_table = {}

    # vm id lookup table for instance_id -> docker node id
    vm_id_table = {}
    # vm ip lookup table for instance_id -> node pvt IP
    vm_ip_table = {}

    # Map Azure VMSS instance IDs to swarm node IDs using NIC and IP address
    nics = network_client.network_interfaces.list_virtual_machine_scale_set_network_interfaces(
                                                            RG_NAME, vmss_name)
    for nic in nics:
        LOG.debug("NIC: {} Primary:{}".format(nic.id, nic.primary))
        if nic.primary:
            for ips in nic.ip_configurations:
                LOG.debug("IP: {} Primary:{}".format(ips.private_ip_address, ips.primary))
                if ips.primary:
                    nic_id_table[nic.id] = ips.private_ip_address

    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, vmss_name)
    for vm in vms:
        LOG.debug("Enumerate NICs of VM: {} in VMSS {}".format(vm.instance_id, vmss_name))
        for nic in vm.network_profile.network_interfaces:
            if nic.id in nic_id_table:
                ip = nic_id_table[nic.id]
                LOG.debug("IP Address of NIC: {}".format(ip))
                vm_tmp_table[ip] = vm.instance_id
                vm_ip_table[vm.instance_id] = ip

    nodes = docker_client.nodes(filters={'role': VMSS_ROLE_MAPPING[vmss_name]})
    for node in nodes:
        node_ip = node['Status']['Addr']
        LOG.debug("Swarm Node ID: {} IP: {}".format(node['ID'], node_ip))
        if node_ip not in vm_tmp_table:
            LOG.error("Node IP {} not found in list of VM IPs".format(node_ip))
            return
        vm_id_table[vm_tmp_table[node_ip]] = node['ID']

    # at this point all sanity checks and metadata gathering should be complete
    # beyond this point it's hard to recover from any missing data/errors
    for vm in vms:
        node_id = vm_id_table[vm.instance_id]
        LOG.info("Accessing VM: {} in VMSS {}".format(vm.instance_id, vmss_name))

        # skip the node we are invoking from for now
        if docker_client.info()['Swarm']['NodeID'] == node_id:
            LOG.info("Skip upgrading of manager node where upgrade is initiated from for now")
            LAST_MANAGER_NODE_ID = docker_client.info()['Swarm']['NodeID']
            LOG.info("LAST MANAGER Node ID: {}".format(LAST_MANAGER_NODE_ID))
            continue

        # demote managers and keep track of new leader
        node_info = docker_client.inspect_node(node_id)
        node_hostname = node_info['Description']['Hostname']
        if node_info['Spec']['Role'] == 'manager':
            try:
                leader = node_info['ManagerStatus']['Leader']
            except KeyError:
                leader = False

            if HAS_DDC:
                checkDDC(docker_client, node_hostname)

            cmdout = subprocess.check_output(["docker", "node", "demote", node_id])
            LOG.info("docker node demote output: {}".format(cmdout))
            sleep(10)

            cmdout = subprocess.check_output(["docker", "node", "ls"])
            LOG.info("docker node ls output: {}".format(cmdout))

            # If node was a leader, update IP of new leader
            # (after demotion of previous leader) in table
            if leader:
                LOG.info("Previous Leader demoted. Update leader IP address")
                leader_ip = get_swarm_leader_ip(docker_client)
                update_leader_tbl(tbl_svc, SWARM_TABLE, LEADER_PARTITION,
                                    LEADER_ROW, leader_ip)

        cmdout = subprocess.check_output(["docker", "node", "rm", "--force", node_id])
        LOG.info("docker node rm output: {}".format(cmdout))

        cmdout = subprocess.check_output(["docker", "node", "ls"])
        LOG.info("docker node ls output: {}".format(cmdout))

        LOG.info("Update OS info started for VMSS node: {}".format(vm.instance_id))
        async_vmss_update = compute_client.virtual_machine_scale_sets.update_instances(
                                            RG_NAME, vmss_name, vm.instance_id)
        wait_with_status(async_vmss_update, "Waiting for VM OS info update to complete ...")
        LOG.info("Update OS info completed for VMSS node: {}".format(vm.instance_id))

        cmdout = subprocess.check_output(["docker", "node", "ls"])
        LOG.info("docker node ls output: {}".format(cmdout))

        LOG.info("Reimage started for VMSS node: {}".format(vm.instance_id))
        async_vmss_update = compute_client.virtual_machine_scale_set_vms.reimage(
                                            RG_NAME, vmss_name, vm.instance_id)
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

        node_ready = False
        while not node_ready:
            LOG.info("Waiting for VMSS node:{} to boot and join back in swarm".format(
                    vm.instance_id))
            sleep(10)
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
        LOG.info("VMSS node:{} successfully connected back to swarm".format(
                vm.instance_id))


def main():

    with open(LOG_CFG_FILE) as log_cfg_file:
        log_cfg = json.load(log_cfg_file)
        logging.config.dictConfig(log_cfg)

    parser = argparse.ArgumentParser(description='Upgrade a Docker4Azure resource group')
    parser.add_argument('template_url', help='New template to upgrade to')

    args = parser.parse_args()

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID,
        resource=RESOURCE_MANAGER_ENDPOINT,
        auth_uri=ACTIVE_DIRECTORY_ENDPOINT
    )

    docker_client = Client(base_url='unix://var/run/docker.sock', version="1.25")

    resource_client = ResourceManagementClient(cred, SUB_ID, api_version='2016-09-01', base_url=RESOURCE_MANAGER_ENDPOINT)
    storage_client = StorageManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)
    compute_client = ComputeManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)
    # the default API version for the REST APIs for Network points to 2016-06-01
    # which does not have several VMSS NIC APIs we need. So specify version here
    network_client = NetworkManagementClient(cred, SUB_ID, api_version='2016-09-01', base_url=RESOURCE_MANAGER_ENDPOINT)

    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}
    tbl_svc = TableService(account_name=SA_NAME, account_key=storage_keys['key1'], endpoint_suffix=STORAGE_ENDPOINT)

    try:
        LOG.info("Validate Template URL to upgrade to")
        validate_template(args.template_url)
    except:
        LOG.error("Template validation failed. Please make sure the template URL has a valid JSON file and is accessible.")
        raise

    qsvc = QueueService(account_name=SA_NAME, account_key=storage_keys['key1'], endpoint_suffix=STORAGE_ENDPOINT)
    # the Upgrade Msg Queue should only exist when an upgrade is in progress
    if qsvc.exists(UPGRADE_MSG_QUEUE):
        LOG.error("Upgrade message queue already exists. Please make sure another upgrade is not in progress.")
        return

    LOG.info("Initiating upgrade. Create queue to prevent another simultaneous upgrade.")
    qsvc.create_queue(UPGRADE_MSG_QUEUE)

    try:
        # Update the resource group template
        LOG.info("Updating Resource Group template. This will take several minutes. You can follow the status of the upgrade below or from the Azure console using the URL below:")
        LOG.info("https://portal.azure.com/#resource/subscriptions/{}/resourceGroups/{}/overview".format(
                SUB_ID, RG_NAME))
        update_deployment_template(args.template_url, resource_client)

        # Update manager nodes (except the one this script is initiated from)
        LOG.info("Starting rolling upgrade of swarm manager nodes. This will take several minutes. You can follow the status of the upgrade below or from the Azure console using the URL below:")
        LOG.info("https://portal.azure.com/#resource/subscriptions/{}/resourceGroups/{}/providers/Microsoft.Compute/virtualMachineScaleSets/{}/overview".format(
                SUB_ID, RG_NAME, MGR_VMSS_NAME))
        update_vmss(MGR_VMSS_NAME, docker_client, compute_client, network_client, tbl_svc)

        # Update worker nodes
        LOG.info("Starting rolling upgrade of swarm worker nodes. This will take several minutes. You can follow the status of the upgrade below or from the Azure console using the URL below:")
        LOG.info("https://portal.azure.com/#resource/subscriptions/{}/resourceGroups/{}/providers/Microsoft.Compute/virtualMachineScaleSets/{}/overview".format(
                SUB_ID, RG_NAME, WRK_VMSS_NAME))
        update_vmss(WRK_VMSS_NAME, docker_client, compute_client, network_client, tbl_svc)

    # remove DTR from last manager node before upgrade
        if HAS_DDC:
            LOG.info("Remove DTR and update last manager")
            update_lastmanager(docker_client)

        LOG.info("The current VM will be rebooted soon for an upgrade. You can follow the status of the upgrade from the Azure console using the URL below:")
        LOG.info("https://portal.azure.com/#resource/subscriptions/{}/resourceGroups/{}/providers/Microsoft.Compute/virtualMachineScaleSets/{}/overview".format(
                SUB_ID, RG_NAME, MGR_VMSS_NAME))
        # Signal another node to upgrade the current node and update leader if necessary
        qsvc.put_message(UPGRADE_MSG_QUEUE, docker_client.info()['Swarm']['NodeID'])

    except:
        LOG.error("The upgrade process encountered errors:", exc_info=True)
        qsvc.delete_queue(UPGRADE_MSG_QUEUE)
        raise

if __name__ == "__main__":
    main()

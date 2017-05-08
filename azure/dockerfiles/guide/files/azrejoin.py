#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
import urllib2
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
from azendpt import *

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']
ROLE = os.environ['ROLE']

RG_NAME = os.environ['GROUP_NAME']
SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']
IP_ADDR = os.environ['PRIVATE_IP']

LOG_CFG_FILE = "/etc/azrejoin_log_cfg.json"
LOG = logging.getLogger("azrejoin")

RESOURCE_MANAGER_ENDPOINT = os.getenv('RESOURCE_MANAGER_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['RESOURCE_MANAGER_ENDPOINT'])
ACTIVE_DIRECTORY_ENDPOINT = os.getenv('ACTIVE_DIRECTORY_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['ACTIVE_DIRECTORY_ENDPOINT'])
STORAGE_ENDPOINT = os.getenv('STORAGE_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['STORAGE_ENDPOINT'])

def rejoin_swarm(leader_ip):
    docker_client = Client(base_url='unix://var/run/docker.sock', version="1.25")
    wrk_token = 0
    token_recvd = False
    while not token_recvd:
        try:
            response = urllib2.urlopen(WRK_TOKEN_ENDPOINT.format(leader_ip))
            wrk_token = response.read()
            token_recvd = True
            LOG.info("Token received:{}".format(wrk_token))
        except urllib2.HTTPError, e:
            LOG.info("HTTPError {} when retrieving token. Retry.".format(str(e.code)))
            sleep(60)
        except urllib2.URLError, e:
            LOG.info("URLError {} when retrieving token. Retry.".format(str(e.reason)))
            sleep(60)
    try:
        docker_client.leave_swarm()
        LOG.info("Left stale swarm that node was attached to")
    except docker.errors.APIError, e:
        LOG.warning("Error when leaving swarm. Okay to continue.")

    LOG.info("Rejoining swarm with leader ip: {}".format(leader_ip))
    docker_client.join_swarm(["{}:{}".format(leader_ip, SWARM_LISTEN_PORT)],
                                wrk_token, SWARM_LISTEN_ADDR)
    LOG.info("Successfully rejoined swarm")

def main():

    # run only in workers for single manager upgrade scenarios
    if ROLE == MGR_ROLE:
        return

    with open(LOG_CFG_FILE) as log_cfg_file:
        log_cfg = json.load(log_cfg_file)
        logging.config.dictConfig(log_cfg)

    LOG.debug("Rejoin listener started")

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


    if not qsvc.exists(REJOIN_MSG_QUEUE):
        LOG.debug("Rejoin message queue does not exist")
        return

    if not tbl_svc.exists(SWARM_TABLE):
        LOG.debug("Swarm table does not exist")
        return

    leader_data = tbl_svc.get_entity(SWARM_TABLE, LEADER_PARTITION, LEADER_ROW)
    leader_ip = leader_data.manager_ip

    metadata = qsvc.get_queue_metadata(REJOIN_MSG_QUEUE)
    count = metadata.approximate_message_count
    if count == 0:
        LOG.debug("Nothing detected in rejoin message queue")
        return

    LOG.info("Process messages in rejoin message queue")
    # backoff unless the msg is destined for self. Otherwise others may get starved
    msgs = qsvc.peek_messages(REJOIN_MSG_QUEUE)
    for msg in msgs:
        node_ip = msg.content
        LOG.info("Check node IP. Obtained Node IP: {}".format(node_ip))
        if node_ip != IP_ADDR:
            LOG.info("Msg destined for another node.")
            return

    # this will be tried every minute by a worker. The above peek will ensure this
    # will be tried by the worker for which the msg is destined. So it may take a worst case
    # of N minutes if there are N workers. Since this code deals only with single
    # manager swarms, N is expected to be low and this won't take too long.
    LOG.info("Message detected for IP Address {}".format(IP_ADDR))
    msgs = qsvc.get_messages(REJOIN_MSG_QUEUE)
    for msg in msgs:
        node_ip = msg.content
        if node_ip != IP_ADDR:
            LOG.info("Message for self no longer available. Try later.")
            return
        LOG.info("Remove message destined for self from rejoin queue")
        qsvc.delete_message(REJOIN_MSG_QUEUE, msg.id, msg.pop_receipt)

    LOG.info("Rejoin the swarm")
    rejoin_swarm(leader_ip)

if __name__ == "__main__":
    main()

#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
import urllib2
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

RG_NAME = os.environ['GROUP_NAME']
SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']
IP_ADDR = os.environ['PRIVATE_IP']


def rejoin_swarm(leader_ip):
    docker_client = Client(base_url='unix://var/run/docker.sock', version="1.25")
    wrk_token = 0
    token_recvd = False
    while not token_recvd:
        try:
            response = urllib2.urlopen(WRK_TOKEN_ENDPOINT.format(leader_ip))
            wrk_token = response.read()
            token_recvd = True
            print("Token received:{}".format(wrk_token))
        except urllib2.HTTPError, e:
            print("HTTPError {} when retrieving token. Retry.".format(str(e.code)))
            sleep(60)
        except urllib2.URLError, e:
            print("URLError {} when retrieving token. Retry.".format(str(e.reason)))
            sleep(60)
    try:
        docker_client.leave_swarm()
        print("Left stale swarm that node was attached to")
    except docker.errors.APIError, e:
        print("Error when leaving swarm. Okay to continue.")

    print("Rejoining swarm with leader ip: {}".format(leader_ip))
    docker_client.join_swarm(["{}:{}".format(leader_ip, SWARM_LISTEN_PORT)],
                                wrk_token, SWARM_LISTEN_ADDR)

def main():

    # run only in workers for single manager upgrade scenarios
    if ROLE == MGR_ROLE:
        return
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


    if not qsvc.exists(REJOIN_MSG_QUEUE):
        return

    if not tbl_svc.exists(SWARM_TABLE):
        return

    leader_data = tbl_svc.get_entity(SWARM_TABLE, LEADER_PARTITION, LEADER_ROW)
    leader_ip = leader_data.manager_ip

    metadata = qsvc.get_queue_metadata(REJOIN_MSG_QUEUE)
    count = metadata.approximate_message_count
    if count == 0:
        print("nothing detected in queue yet")
        return

    # backoff unless the msg is destined for self. Otherwise others may get starved
    msgs = qsvc.peek_messages(REJOIN_MSG_QUEUE)
    for msg in msgs:
        node_ip = msg.content
        if node_ip != IP_ADDR:
            return

    # this will be tried every minute by a worker. The above peek will ensure this
    # will be tried by the worker for which the msg is destined. So it may take a worst case
    # of N minutes if there are N workers. Since this code deals only with single
    # manager swarms, N is expected to be low and this won't take too long.
    print("Message detected for IP Address {}".format(IP_ADDR))
    msgs = qsvc.get_messages(REJOIN_MSG_QUEUE)
    for msg in msgs:
        node_ip = msg.content
        if node_ip != IP_ADDR:
            return
        qsvc.delete_message(REJOIN_MSG_QUEUE, msg.id, msg.pop_receipt)

    rejoin_swarm(leader_ip)

if __name__ == "__main__":
    main()

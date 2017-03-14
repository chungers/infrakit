#!/usr/bin/env python

import os
import argparse
import sys
from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.storage import StorageManagementClient
from azure.mgmt.compute import ComputeManagementClient
from azure.mgmt.storage.models import StorageAccountCreateParameters
from azure.storage.table import TableService, Entity

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']
RG_NAME = os.environ['GROUP_NAME']
SA_DTR_NAME = os.environ['DTR_STORAGE_ACCOUNT']

#Azure Table Name to store DTR replicas and serialize flag
DTR_TBL_NAME = 'dtrtable'

VMSS_MGR = 'swarm-manager-vmss'
VMSS_WRK = 'swarm-worker-vmss'


def get_mgr_nodes():
    global MGR_VMSS_NAME

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID
    )
    compute_client = ComputeManagementClient(cred, SUB_ID)
    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, VMSS_MGR)
    mgr_nodes = 0
    for vm in vms:
        mgr_nodes += 1
    print("{}".format(mgr_nodes))
    
def get_wrk_nodes():
    global WRK_VMSS_NAME

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID
    )
    compute_client = ComputeManagementClient(cred, SUB_ID)
    vms = compute_client.virtual_machine_scale_set_vms.list(RG_NAME, VMSS_WRK)
    wrk_nodes = 0
    for vm in vms:
        wrk_nodes += 1
    print("{}".format(wrk_nodes))

def get_storage_key():
    global SUB_ID, TENANT_ID, APP_ID, APP_SECRET, RG_NAME, SA_DTR_NAME
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID
    )

    resource_client = ResourceManagementClient(cred, SUB_ID)
    storage_client = StorageManagementClient(cred, SUB_ID)

    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_DTR_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}

    return storage_keys['key1']

def get_key(sa_key):
	print("{}".format(sa_key))

def get_replica_ids(sa_key):
    global SA_DTR_NAME, DTR_TBL_NAME
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key)
    if not tbl_svc.exists(DTR_TBL_NAME):
        return False
    try:
        replicaids = tbl_svc.query_entities(DTR_TBL_NAME, filter="PartitionKey eq 'dtrreplicas'")
	for replicaid in replicaids:
		print("{}".format(replicaid.replica_id))
        return replicaids
    except:
        return False

def get_desc(sa_key):
    global SA_DTR_NAME, DTR_TBL_NAME
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key)
    if not tbl_svc.exists(DTR_TBL_NAME):
        return False
    try:
        replicaids = tbl_svc.query_entities(DTR_TBL_NAME, filter="PartitionKey eq 'dtrreplicas'", select='description')
	for replicaid in replicaids:
		if replicaid.description == "upgrade in progress":
                	print("{}".format(replicaid.description))
			return replicaid.description
    except:
        return False

def print_id(sa_key):
    global SA_DTR_NAME, DTR_TBL_NAME
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key)
    if not tbl_svc.exists(DTR_TBL_NAME):
        return False
    try:
        replicaid = tbl_svc.get_entity(DTR_TBL_NAME, 'dtrseqid', '1')
        print("{}".format(replicaid.replica_id))
	return True
    except:
        return False


def insert_id(sa_key, replica_id, node_name, description):
    global  DTR_TBL_NAME, SA_DTR_NAME
    print ("replica id:{}".format(replica_id))
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key)
    dtr_id = {'PartitionKey': 'dtrreplicas', 'RowKey': replica_id, 'replica_id': replica_id, 'node_name': node_name, 'description':description}
    try:
        # this upsert operation should always succeed
        tbl_svc.insert_or_replace_entity(DTR_TBL_NAME, dtr_id)
        print("successfully inserted/replaced replica-id {}".format(replica_id))
        return True
    except:
        print("exception while inserting replica-id")
        return False

def add_id(sa_key, id):
    global  DTR_TBL_NAME, SA_DTR_NAME
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key)
    try:
        # this upsert operation should always succeed
    	dtr_id = {'PartitionKey': 'dtrseqid', 'RowKey': '1', 'replica_id': id}
        tbl_svc.insert_or_replace_entity(DTR_TBL_NAME, dtr_id)
        print("successfully inserted/replaced dtr ID {}".format(id))
        return True
    except:
        print("exception while inserting DTR ID")
        return False

def delete_id(sa_key, replica_id):
    global DTR_TBL_NAME, SA_DTR_NAME
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key)
    dtr_id = {'PartitionKey': 'dtrreplicas', 'RowKey': replica_id}
    try:
        # this upsert operation should always succeed
        tbl_svc.delete_entity(DTR_TBL_NAME, dtr_id)
        print("successfully deleted replica-id")
        return True
    except:
        print("exception while deleting replica-id")
        return False


def create_table(sa_key):
    global DTR_TBL_NAME, SA_DTR_NAME
    tbl_svc = TableService(account_name=SA_DTR_NAME, account_key=sa_key)
    try:
        # this will succeed only once for a given table name on a storage account
        tbl_svc.create_table(DTR_TBL_NAME, fail_on_exist=True)
        print("successfully created table")
        return True
    except:
        print("exception while creating table")
        return False

def main():

    parser = argparse.ArgumentParser(description='Tool to store Docker Trusted Registry Replica info in Azure Tables')
    subparsers = parser.add_subparsers(help='commands', dest='action')
    create_table_parser = subparsers.add_parser('create-table', help='Create table specified in env var DTR_TBL_NAME')
    get_key_parser = subparsers.add_parser('get-key', help='Get DTR storage key')
    get_id_parser = subparsers.add_parser('get-id', help='Get DTR Replica info from table specified in env var DTR_TBL_NAME')
    get_ids_parser = subparsers.add_parser('get-ids', help='Get DTR Replica info from table specified in env var DTR_TBL_NAME')
    insert_id_parser = subparsers.add_parser('insert-id', help='Insert DTR Replica info to table specified in env var DTR_TBL_NAME')
    insert_id_parser.add_argument('id', help='Replica Id of the DTR')
    insert_id_parser.add_argument('name', help='Hostname of the manager node')
    insert_id_parser.add_argument('desc', help='description for replica -- initial install/upgrade')
    add_id_parser = subparsers.add_parser('add-id', help='Insert DTR SEQ ID to table specified in env var DTR_TBL_NAME')
    add_id_parser.add_argument('id', help='update DTR SEQ ID to table specified in env var DTR_TBL_NAME')
    get_desc_parser = subparsers.add_parser('get-desc', help='Get desc from table specified in env var DTR_TBL_NAME')
    delete_id_parser = subparsers.add_parser('delete-id', help='Insert DTR Replica info to table specified in env var DTR_TBL_NAME')
    delete_id_parser.add_argument('id', help='Replica Id of the leader')
    get_mgr_nodes_parser = subparsers.add_parser('get-mgr-nodes', help='Get VMSS manager node count')
    get_wrk_nodes_parser = subparsers.add_parser('get-wrk-nodes', help='Get VMSS worker node count')

    args = parser.parse_args()

    key=get_storage_key()

    if args.action == 'create-table':
        if not create_table(key):
            sys.exit(1)
    elif args.action == 'get-key':
        get_key(key)
    elif args.action == 'insert-id':
        insert_id(key, args.id, args.name, args.desc)
    elif args.action == 'add-id':
        add_id(key, args.id)
    elif args.action == 'delete-id':
        delete_id(key, args.id)
    elif args.action == 'get-id':
        print_id(key)
    elif args.action == 'get-desc':
        get_desc(key)
    elif args.action == 'get-ids':
        get_replica_ids(key)
    elif args.action == 'get-mgr-nodes':
        get_mgr_nodes()
    elif args.action == 'get-wrk-nodes':
        get_wrk_nodes()
    else:
        parser.print_usage()

if __name__ == "__main__":
    main()


import os
import argparse
import sys
from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.storage import StorageManagementClient
from azure.mgmt.storage.models import StorageAccountCreateParameters
from azure.storage.table import TableService, Entity

PARTITION_NAME = 'tokens'
ROW_ID = '1'

def get_storage_key():
    sub_id = os.environ['ACCOUNT_ID']
    cred = ServicePrincipalCredentials(
        client_id=os.environ['APP_ID'],
        secret=os.environ['APP_SECRET'],
        tenant=os.environ['TENANT_ID']
    )

    resource_client = ResourceManagementClient(cred, sub_id)
    storage_client = StorageManagementClient(cred, sub_id)

    rg_name = os.environ['GROUP_NAME']
    sa_name = os.environ['SWARM_INFO_STORAGE_ACCOUNT']

    storage_keys = storage_client.storage_accounts.list_keys(rg_name, sa_name)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}

    return storage_keys['key1']

def get_tokens(sa_key):
    global PARTITION_NAME
    global ROW_ID
    tbl_name = os.environ['SWARM_INFO_TABLE']
    sa_name = os.environ['SWARM_INFO_STORAGE_ACCOUNT']
    tbl_svc = TableService(account_name=sa_name, account_key=sa_key)
    if not tbl_svc.exists(tbl_name):
        return False
    try:
        token = tbl_svc.get_entity(tbl_name, PARTITION_NAME, ROW_ID)
        print '{}|{}|{}'.format(token.manager_ip, token.manager_token, token.worker_token)
        return True
    except:
        return False

def insert_tokens(sa_key, manager_ip, manager_token, worker_token):
    global PARTITION_NAME
    global ROW_ID
    tbl_name = os.environ['SWARM_INFO_TABLE']
    sa_name = os.environ['SWARM_INFO_STORAGE_ACCOUNT']
    tbl_svc = TableService(account_name=sa_name, account_key=sa_key)
    token = {'PartitionKey': PARTITION_NAME, 'RowKey': ROW_ID, 'manager_ip': manager_ip, 'manager_token': manager_token, 'worker_token': worker_token}
    try:
        tbl_svc.insert_entity(tbl_name, token)
        print "successfully inserted tokens"
        return True
    except:
        print "exception while inserting tokens"
        return False

def create_table(sa_key):
    tbl_name = os.environ['SWARM_INFO_TABLE']
    sa_name = os.environ['SWARM_INFO_STORAGE_ACCOUNT']
    print "sa_name: ", sa_name, "tbl_name ", tbl_name
    tbl_svc = TableService(account_name=sa_name, account_key=sa_key)
    try:
        tbl_svc.create_table(tbl_name, fail_on_exist=True)
        print "successfully created table"
        return True
    except:
        print "exception while creating table"
        return False

def main():

    parser = argparse.ArgumentParser(description='Tool to store Docker Swarm info in Azure Tables')
    subparsers = parser.add_subparsers(help='commands', dest='action')
    create_table_parser = subparsers.add_parser('create-table', help='Create table specified in env var AZURE_TBL_NAME')
    get_tokens_parser = subparsers.add_parser('get-tokens', help='Get swarm info from table specified in env var AZURE_TBL_NAME')
    insert_tokens_parser = subparsers.add_parser('insert-tokens', help='Insert swarm info to table specified in env var AZURE_TBL_NAME')
    insert_tokens_parser.add_argument('ip', help='IP address of the primary swarm manager')
    insert_tokens_parser.add_argument('manager_token', help='Manager Token for the swarm')
    insert_tokens_parser.add_argument('worker_token', help='Worker token for the swarm')

    args = parser.parse_args()

    key=get_storage_key()

    if args.action == 'create-table':
        if not create_table(key):
            sys.exit(1)
    elif args.action == 'get-tokens':
        get_tokens(key)
    elif args.action == 'insert-tokens':
        insert_tokens(key, args.ip, args.manager_token, args.worker_token)
    else:
        parser.print_usage()

if __name__ == "__main__":
    main()

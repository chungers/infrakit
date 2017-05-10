#!/usr/bin/env python
import os
import argparse
import sys
from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.storage import StorageManagementClient
from azure.mgmt.storage.models import StorageAccountCreateParameters
from azure.storage.table import TableService, Entity
from azendpt import AZURE_PLATFORMS, AZURE_DEFAULT_ENV

SUB_ID = os.environ['ACCOUNT_ID']
APP_ID = os.environ['APP_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_SECRET = os.environ['APP_SECRET']
RG_NAME = os.environ['GROUP_NAME']
SA_NAME = os.environ['SWARM_INFO_STORAGE_ACCOUNT']

RESOURCE_MANAGER_ENDPOINT = os.getenv('RESOURCE_MANAGER_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['RESOURCE_MANAGER_ENDPOINT'])
ACTIVE_DIRECTORY_ENDPOINT = os.getenv('ACTIVE_DIRECTORY_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['ACTIVE_DIRECTORY_ENDPOINT'])
STORAGE_ENDPOINT = os.getenv('STORAGE_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['STORAGE_ENDPOINT'])

def get_storage_key():
    global SUB_ID, TENANT_ID, APP_ID, APP_SECRET, RG_NAME, SA_NAME
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID,
        resource=RESOURCE_MANAGER_ENDPOINT,
        auth_uri=ACTIVE_DIRECTORY_ENDPOINT
    )

    resource_client = ResourceManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)
    storage_client = StorageManagementClient(cred, SUB_ID, base_url=RESOURCE_MANAGER_ENDPOINT)

    storage_keys = storage_client.storage_accounts.list_keys(RG_NAME, SA_NAME)
    storage_keys = {v.key_name: v.value for v in storage_keys.keys}

    return storage_keys['key1']

def main():
    key = get_storage_key()
    print(u'{}'.format(key))

if __name__ == "__main__":
    main()

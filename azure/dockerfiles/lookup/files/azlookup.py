#!/usr/bin/env python

import os
import json
import argparse
import sys
import subprocess
import pytz
import urllib2
from datetime import datetime
from time import sleep
from docker import Client
from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.storage import StorageManagementClient
from azure.mgmt.compute import ComputeManagementClient
from azure.mgmt.network import NetworkManagementClient
from azure.mgmt.resource import ResourceManagementClient
from azure.mgmt.resource.resources.models import DeploymentMode
from azure.mgmt.storage.models import StorageAccountCreateParameters
from azure.storage.table import TableService, Entity
from azure.storage.queue import QueueService
from azendpt import AZURE_PLATFORMS, AZURE_DEFAULT_ENV
from aztags import get_tag_value

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']

RG_NAME = os.environ['GROUP_NAME']

RESOURCE_MANAGER_ENDPOINT = os.getenv('RESOURCE_MANAGER_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['RESOURCE_MANAGER_ENDPOINT'])
ACTIVE_DIRECTORY_ENDPOINT = os.getenv('ACTIVE_DIRECTORY_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['ACTIVE_DIRECTORY_ENDPOINT'])


def get_deployment_parameter(resource_client, parameter_name):
	
    deployments = resource_client.deployments.list(RG_NAME)
    latest_timestamp = datetime.min.replace(tzinfo=pytz.UTC)
    latest_deployment = None

    for deployment in deployments:
        if deployment.properties is None:
            continue
        state = deployment.properties.provisioning_state
        if state != "Succeeded":
            continue

        timestamp = deployment.properties.timestamp
        if timestamp > latest_timestamp:
            latest_timestamp = timestamp
            latest_deployment = deployment

    if latest_deployment is None:
        raise RuntimeError(u"No successful deployment found for {}".format(RG_NAME))

    if latest_deployment.properties.parameters is None:
        raise RuntimeError(u"Parameter Link Not Supported")

    return latest_deployment.properties.parameters[parameter_name]['value']

def main():

    parser = argparse.ArgumentParser(description='Fetch tag/parameter value of any resource in a resource group')
    parser.add_argument('--tag', action='store_true', default=False, help='Type of the value to fetch (default is parameter, unless the flag is set)')
    parser.add_argument('name', help='Name of the tag/parameter whose value to fetch')

    args = parser.parse_args()

    if not args.name:
        raise Exception("No resource name was defined")

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID,
        resource=RESOURCE_MANAGER_ENDPOINT,
        auth_uri=ACTIVE_DIRECTORY_ENDPOINT
    )

    resource_client = ResourceManagementClient(cred, SUB_ID, api_version='2016-09-01', base_url=RESOURCE_MANAGER_ENDPOINT)
    if args.tag:
        print(get_tag_value(resource_client, args.name))
    else:
        print(get_deployment_parameter(resource_client, args.name))

if __name__ == "__main__":
    main()

#!/usr/bin/env python

import os
import argparse
import sys

from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.resource import ResourceManagementClient

SUB_ID = os.environ['ACCOUNT_ID']
TENANT_ID = os.environ['TENANT_ID']
APP_ID = os.environ['APP_ID']
APP_SECRET = os.environ['APP_SECRET']

RG_NAME = os.environ['GROUP_NAME']


def get_tag_value(resource_client, tag_name):
    for item in resource_client.resource_groups.list_resources(RG_NAME):
        if tag_name in item.tags:
            return item.tags[tag_name]
    raise KeyError(tag_name + " Not found in any resource")

def main():

    parser = argparse.ArgumentParser(description='Fetch tag value of any resource in a resource group')
    parser.add_argument('tag_name', help='Name of the tag whose value to fetch')

    args = parser.parse_args()

    # init various Azure API clients using credentials
    cred = ServicePrincipalCredentials(
        client_id=APP_ID,
        secret=APP_SECRET,
        tenant=TENANT_ID
    )

    resource_client = ResourceManagementClient(cred, SUB_ID)
    print(get_tag_value(resource_client, args.tag_name))

if __name__ == "__main__":
    main()

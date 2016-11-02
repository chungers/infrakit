# Copyright 2016 Docker Inc. All rights reserved.

"""Swarm manager."""

COMPUTE_URL_BASE = 'https://www.googleapis.com/compute/v1'

def GenerateConfig(context):
  script = r"""
#!/bin/bash

set -x
service docker start

docker swarm init --advertise-addr ens4:2377 --listen-addr ens4:2377

TOKEN_64=$(docker swarm join-token worker -q | base64 -w0 -)
echo ${TOKEN_64}

PROJECT=$(curl -s http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
echo ${PROJECT}

ACCESS_TOKEN=$(curl -s http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token -H "Metadata-Flavor: Google" | jq -r ".access_token")
echo ${ACCESS_TOKEN}

curl -sX PUT -H "Content-Type: application/json" -d "{\"value\":\"${TOKEN_64}\"}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/token -H "Authorization":"Bearer ${ACCESS_TOKEN}"

TOKEN=$(curl -sSL https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/token -H "Authorization":"Bearer ${ACCESS_TOKEN}" | jq -r ".value" | base64 -d -)
echo ${TOKEN}
"""

  outputs = [{
      'name': 'internalIP',
      'value': '$(ref.' + context.env['name'] + '.networkInterfaces[0].networkIP)'
  }]

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.instance',
      'properties': {
          'zone': context.properties['zone'],
          'tags': {
              'items': ['swarm', 'swarm-manager']
          },
          'machineType': '/'.join([COMPUTE_URL_BASE, 'projects', context.env['project'],
                                  'zones', context.properties['zone'],
                                  'machineTypes', context.properties['machineType']]),
          'disks': [{
              'deviceName': 'boot',
              'type': 'PERSISTENT',
              'boot': True,
              'autoDelete': True,
              'initializeParams': {
                  'sourceImage': '/'.join([COMPUTE_URL_BASE, 'projects',
                                          context.env['project'], 'global',
                                          'images', 'docker2'])
              }
          }],
          'networkInterfaces': [{
              'network': '$(ref.' + context.properties['network'] + '.selfLink)',
              'accessConfigs': [{
                  'name': 'External NAT',
                  'type': 'ONE_TO_ONE_NAT'
              }]
          }],
          'metadata': {
              'items': [{
                  'key': 'startup-script',
                  'value': script
              }]
          },
          'serviceAccounts': [{
              'scopes': [
                  'https://www.googleapis.com/auth/cloudruntimeconfig'
              ]
          }]
      }
  }]
  return {'resources': resources, 'outputs': outputs}

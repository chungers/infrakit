# Copyright 2016 Docker Inc. All rights reserved.

"""Swarm manager."""

def GenerateConfig(context):
  project = context.env['project']
  zone = context.properties['zone']
  machineType = context.properties['machineType']
  image = context.properties['image']
  network = '$(ref.' + context.properties['network'] + '.selfLink)'

  script = r"""
#!/bin/bash
set -x

service docker start
docker swarm init --advertise-addr ens4:2377 --listen-addr ens4:2377

TOKEN=$(docker swarm join-token worker -q)
PROJECT=$(curl -s http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
ACCESS_TOKEN=$(curl -s http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token -H "Metadata-Flavor: Google" | jq -r ".access_token")

curl -s -X PUT -H "Content-Type: application/json" -d "{\"text\":\"${TOKEN}\"}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/token -H "Authorization":"Bearer ${ACCESS_TOKEN}"
"""

  outputs = [{
      'name': 'internalIP',
      'value': '$(ref.' + context.env['name'] + '.networkInterfaces[0].networkIP)'
  }]

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.instance',
      'properties': {
          'zone': zone,
          'tags': {
              'items': ['swarm', 'swarm-manager']
          },
          'machineType': '/'.join(['projects', project,
                                  'zones', zone,
                                  'machineTypes', machineType]),
          'disks': [{
              'deviceName': 'boot',
              'type': 'PERSISTENT',
              'boot': True,
              'autoDelete': True,
              'initializeParams': {
                  'sourceImage': image
              }
          }],
          'networkInterfaces': [{
              'network': network,
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

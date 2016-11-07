# Copyright 2016 Docker Inc. All rights reserved.

"""Worker Instance Template."""

COMPUTE_URL_BASE = 'https://www.googleapis.com/compute/v1'

def GenerateConfig(context):
  project = context.env['project']
  zone = context.properties['zone']
  machineType = context.properties['machineType']
  managerIP = context.properties['managerIP']
  network = '$(ref.' + context.properties['network'] + '.selfLink)'

  script = r"""
#!/bin/bash
set -x

service docker start

PROJECT=$(curl -s http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
ACCESS_TOKEN=$(curl -s http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token -H "Metadata-Flavor: Google" | jq -r ".access_token")

for i in $(seq 1 300); do
    TOKEN=$(curl -sSL "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/token" -H "Authorization":"Bearer ${ACCESS_TOKEN}" | jq -r ".text")
    if [ "${TOKEN}" != "" ]; then
        docker swarm join --token "${TOKEN}" """ + managerIP + r""" --advertise-addr ens4:2377 --listen-addr ens4:2377
        break
    fi

    sleep 1
done
"""

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.instanceTemplate',
      'properties': {
          'properties': {
              'zone': zone,
              'machineType': machineType,
              'tags': {
                  'items': ['swarm', 'swarm-manager']
              },
              'disks': [{
                  'deviceName': 'boot',
                  'type': 'PERSISTENT',
                  'boot': True,
                  'autoDelete': True,
                  'initializeParams': {
                      'sourceImage': '/'.join(['projects', project,
                                              'global',
                                              'images', 'docker2'])
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
              'scheduling': {
                  'preemptible': False,
                  'onHostMaintenance': 'TERMINATE',
                  'automaticRestart': False
              },
              'serviceAccounts': [{
                  'scopes': [
                      'https://www.googleapis.com/auth/cloudruntimeconfig'
                  ]
              }]
          }
      }
  }]
  return {'resources': resources}

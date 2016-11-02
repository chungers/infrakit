# Copyright 2016 Docker Inc. All rights reserved.

"""Worker Instance Template."""

COMPUTE_URL_BASE = 'https://www.googleapis.com/compute/v1'

def GenerateConfig(context):
  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.instanceTemplate',
      'properties': {
          'properties': {
              'zone': context.properties['zone'],
              'machineType': context.properties['machineType'],
              'tags': {
                  'items': ['swarm', 'swarm-manager']
              },
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
                      'value':
    r"""
    #!/bin/bash

    set -x
    service docker start

    PROJECT=$(curl -s http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
    echo ${PROJECT}

    ACCESS_TOKEN=$(curl -s http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token -H "Metadata-Flavor: Google" | jq -r ".access_token")
    echo ${ACCESS_TOKEN}

    for i in $(seq 1 300); do
      TOKEN_64=$(curl -sSL "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/token" -H "Authorization":"Bearer ${ACCESS_TOKEN}" | jq -r ".value")
      if [ "${TOKEN_64}" != "" ]; then
        TOKEN=$(echo "${TOKEN_64}" | base64 -d -)
        echo "${TOKEN}"
        docker swarm join --token "${TOKEN}" """ + context.properties['managerIP'] + r""" --advertise-addr ens4:2377 --listen-addr ens4:2377
        break
      fi
      sleep 1
    done
    """
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

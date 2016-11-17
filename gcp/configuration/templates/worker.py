# Copyright 2016 Docker Inc. All rights reserved.

"""Worker Instance Template."""

def GenerateConfig(context):
  project = context.env['project']
  zone = context.properties['zone']
  machineType = context.properties['machineType']
  preemptible = context.properties['preemptible']
  image = context.properties['image']
  network = '$(ref.' + context.properties['network'] + '.selfLink)'

  script = r"""
#!/bin/bash
set -x

service docker start

PROJECT=$(curl -s http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
ACCESS_TOKEN=$(curl -s http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token -H "Metadata-Flavor: Google" | jq -r ".access_token")

for i in $(seq 1 300); do
    LEADER_IP=$(curl -sSL "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/leader-ip" -H "Authorization":"Bearer ${ACCESS_TOKEN}" | jq -r ".text // empty")
    if [ ! -z "${LEADER_IP}" ]; then
        TOKEN=$(curl -sSL "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/worker-token" -H "Authorization":"Bearer ${ACCESS_TOKEN}" | jq -r ".text // empty")
        docker swarm join --token "${TOKEN}" "${LEADER_IP}"
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
              'scheduling': {
                  'preemptible': preemptible,
                  'onHostMaintenance': 'TERMINATE',
                  'automaticRestart': False
              },
              'serviceAccounts': [{
                  'scopes': [
                      'https://www.googleapis.com/auth/cloudruntimeconfig',
                      'https://www.googleapis.com/auth/logging.write'
                  ]
              }]
          }
      }
  }]
  return {'resources': resources}

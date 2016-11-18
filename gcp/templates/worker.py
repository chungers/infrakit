# Copyright 2016 Docker Inc. All rights reserved.

"""Worker Instance Template."""

def GenerateConfig(context):
  project = context.env['project']
  zone = context.properties['zone']
  machineType = context.properties['machineType']
  preemptible = context.properties['preemptible']
  image = context.properties['image']
  network = context.properties['network']
  config = context.properties['config']

  script = r"""
#!/bin/bash
set -x

function metadata {
    curl -s "http://metadata.google.internal/computeMetadata/v1/$1" \
        -H "Metadata-Flavor: Google"
}

function get-val {
    PROJECT=$(metadata project/project-id)
    AUTH=$(metadata instance/service-accounts/default/token | jq -r ".access_token")

    curl -s "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/$1" \
        -H "Authorization: Bearer ${AUTH}" | jq -r ".text // empty"
}

service docker start

echo "I'm a worker"

while [ -z "$(get-val worker-token)" ]; do
    sleep 1
done

docker swarm join --token "$(get-val worker-token)" "$(get-val leader-ip)"

exit 0
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

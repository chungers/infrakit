# Copyright 2016 Docker Inc. All rights reserved.

"""Swarm manager."""

def GenerateConfig(context):
  project = context.env['project']
  zone = context.properties['zone']
  machineType = context.properties['machineType']
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

function set-val {
    PROJECT=$(metadata project/project-id)
    AUTH=$(metadata instance/service-accounts/default/token | jq -r ".access_token")

    curl -f -X POST "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${AUTH}" \
        -d "{'name':'projects/${PROJECT}/configs/swarm-config/variables/$1','text':'$2'}"
}

service docker start
docker swarm init --advertise-addr eth0:2377 --listen-addr eth0:2377

set-val leader-ip $(metadata instance/network-interfaces/0/ip)
if [ $? -eq 0 ]; then
    echo "I'm a leader"

    set-val worker-token $(docker swarm join-token worker -q)
    set-val manager-token $(docker swarm join-token manager -q)

    set-val leader-name $(hostname)
    set-val project $(metadata project/project-id)
    set-val zone $(metadata instance/zone)

    exit 0
fi

echo "I'm not a leader"

while [ -z "$(get-val manager-token)" ]; do
    sleep 1
done

docker swarm leave --force
docker swarm join --token "$(get-val manager-token)" "$(get-val leader-ip)" --advertise-addr eth0:2377 --listen-addr eth0:2377

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
                  'preemptible': False,
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

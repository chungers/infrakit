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
docker swarm init --advertise-addr eth0:2377 --listen-addr eth0:2377

PROJECT=$(curl -s http://metadata.google.internal/computeMetadata/v1/project/project-id -H "Metadata-Flavor: Google")
ACCESS_TOKEN=$(curl -s http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token -H "Metadata-Flavor: Google" | jq -r ".access_token")

LEADER_IP=$(curl -s http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip -H "Metadata-Flavor: Google")
curl -f -s -X POST -H "Content-Type: application/json" -d "{'name':'projects/${PROJECT}/configs/swarm-config/variables/leader-ip','text':'${LEADER_IP}'}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables -H "Authorization":"Bearer ${ACCESS_TOKEN}"

if [ $? -eq 0 ]; then
    echo "I'm a leader"

    WORKER_TOKEN=$(docker swarm join-token worker -q)
    curl -s -X POST -H "Content-Type: application/json" -d "{'name':'projects/${PROJECT}/configs/swarm-config/variables/worker-token','text':'${WORKER_TOKEN}'}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables -H "Authorization":"Bearer ${ACCESS_TOKEN}"

    MANAGER_TOKEN=$(docker swarm join-token manager -q)
    curl -s -X POST -H "Content-Type: application/json" -d "{'name':'projects/${PROJECT}/configs/swarm-config/variables/manager-token','text':'${MANAGER_TOKEN}'}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables -H "Authorization":"Bearer ${ACCESS_TOKEN}"
else
    echo "I'm not a leader"

    docker swarm leave --force

    for i in $(seq 1 300); do
        LEADER_IP=$(curl -sSL "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/leader-ip" -H "Authorization":"Bearer ${ACCESS_TOKEN}" | jq -r ".text // empty")
        if [ ! -z "${LEADER_IP}" ]; then
            TOKEN=$(curl -sSL "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/manager-token" -H "Authorization":"Bearer ${ACCESS_TOKEN}" | jq -r ".text // empty")
            docker swarm join --token "${TOKEN}" "${LEADER_IP}" --advertise-addr eth0:2377 --listen-addr eth0:2377
            break
        fi

        sleep 1
    done
fi
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

# Copyright 2016 Docker Inc. All rights reserved.

"""Initial leader."""

def GenerateConfig(context):
  zone = context.properties['zone']
  machineType = context.properties['machineType']
  image = context.properties['image']
  network = context.properties['network']
  managerCount = context.properties['managerCount']
  workerCount = context.properties['workerCount']
  managerIds = '[%s]' % ','.join('"manager-%d"' % i for i in range(1,int(managerCount)+1))

  managerScript = r"""
#!/bin/bash
set -ex

download() {
  curl -s -o /tmp/$1 https://storage.googleapis.com/docker-infrakit/$1
  chmod u+x /tmp/$1
}

runPlugin() {
  download ${1}
  nohup /tmp/"$@" --log 5 >/tmp/log-${1} 2>&1 &
  sleep 1
}

service docker start

runPlugin infrakit-flavor-combo
runPlugin infrakit-flavor-swarm
runPlugin infrakit-flavor-vanilla
runPlugin infrakit-group-default --name group-stateless
runPlugin infrakit-instance-gcp
runPlugin infrakit-manager swarm --proxy-for-group group-stateless --name group
"""

  leaderScript = managerScript + r"""
docker swarm init --advertise-addr eth0:2377 --listen-addr eth0:2377

curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/managersJson > /tmp/managers.json
curl -s -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/workersJson > /tmp/workers.json

download "infrakit"
for i in $(seq 1 60); do /tmp/infrakit group commit /tmp/managers.json && break || sleep 1; done
for i in $(seq 1 60); do /tmp/infrakit group commit /tmp/workers.json && break || sleep 1; done
"""

  infraKitJson = r"""
{
  "ID": "%s",
  "Properties": {
    "Allocation": {
      "%s": %s
    },
    "Instance": {
      "Plugin": "instance-gcp",
      "Properties": {
        "MachineType": "%s",
        "Network": "%s",
        "NamePrefix": "%s",
        "DiskSizeMb": %d,
        "DiskImage": "%s",
        "TargetPool": "%s",
        "Tags": ["swarm"],
        "Scopes": [
          "https://www.googleapis.com/auth/cloudruntimeconfig",
          "https://www.googleapis.com/auth/logging.write"
        ]
      }
    },
    "Flavor": {
      "Plugin": "flavor-combo",
      "Properties": {
        "Flavors": [
          {
            "Plugin": "flavor-vanilla",
            "Properties": {
              "Init": [%s]
            }
          },
          {
            "Plugin": "flavor-swarm",
            "Properties": {
              "Type": "%s",
              "DockerRestartCommand": "service docker restart"
            }
          }
        ]
      }
    }
  }
}
"""

  managerScriptLines = ','.join('"%s"' % (line.replace('"', '\\"')) for line in managerScript.split('\n'))
  managersJson = infraKitJson % ("managers", "LogicalIDS", managerIds, machineType, network, "manager", 10, "docker", "docker-pool", managerScriptLines, "manager")
  workersJson = infraKitJson % ("workers", "Size", workerCount, machineType, network, "worker", 10, "docker", "docker-pool", "", "worker")

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.instance',
      'properties': {
          'zone': zone,
          'machineType': 'zones/' + zone + '/machineTypes/' + machineType,
          'tags': ['swarm'],
          'tags': {
              'items': ['swarm']
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
                  'value': leaderScript
              }, {
                  'key': 'managersJson',
                  'value': managersJson
              }, {
                  'key': 'workersJson',
                  'value': workersJson
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
                  'https://www.googleapis.com/auth/logging.write',
                  'https://www.googleapis.com/auth/compute'
              ]
          }]
      }
  }]
  return {'resources': resources}

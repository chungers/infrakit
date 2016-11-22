# Copyright 2016 Docker Inc. All rights reserved.

"""Creates the Swarm."""

def GenerateConfig(context):
  zone = context.properties['zone']
  managerCount = context.properties['managerCount']
  managerMachineType = context.properties['managerMachineType']
  workerCount = context.properties['workerCount']
  workerMachineType = context.properties['workerMachineType']
  preemptible = context.properties['preemptible']

  resources = [{
      'name': 'docker',
      'type': 'templates/disk-image.py'
  }, {
      'name': 'manager',
      'type': 'templates/manager.py',
      'properties': {
          'zone': zone,
          'machineType': managerMachineType,
          'image': '$(ref.docker.selfLink)',
          'network': 'swarm-network'
      }
  }, {
      'name': 'managers',
      'type': 'templates/managers.py',
      'properties': {
          'zone': zone,
          'template': '$(ref.manager.name)',
          'size': managerCount
      }
  }, {
      'name': 'worker',
      'type': 'templates/worker.py',
      'properties': {
          'zone': zone,
          'machineType': workerMachineType,
          'preemptible': preemptible,
          'image': '$(ref.docker.selfLink)',
          'network': 'swarm-network',
      }
  }, {
      'name': 'workers',
      'type': 'templates/workers.py',
      'properties': {
          'zone': zone,
          'template': '$(ref.worker.name)',
          'size': workerCount
      }
  }, {
      'name': 'swarm-network',
      'type': 'templates/network.py'
  }, {
      'name': 'firewall-rules',
      'type': 'templates/firewall.py',
      'properties': {
          'network': 'swarm-network'
      }
  }]
  return {'resources': resources}

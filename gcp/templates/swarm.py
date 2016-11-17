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
      'type': 'disk-image.py'
  }, {
      'name': 'manager',
      'type': 'manager.py',
      'properties': {
          'zone': zone,
          'machineType': managerMachineType,
          'image': '$(ref.docker.selfLink)',
          'network': '$(ref.swarm-network.selfLink)',
          'config': '$(ref.swarm-config.selfLink)'
      }
  }, {
      'name': 'managers',
      'type': 'managers.py',
      'properties': {
          'zone': zone,
          'template': '$(ref.manager.name)',
          'size': managerCount
      }
  }, {
      'name': 'worker',
      'type': 'worker.py',
      'properties': {
          'zone': zone,
          'machineType': workerMachineType,
          'preemptible': preemptible,
          'image': '$(ref.docker.selfLink)',
          'network': '$(ref.swarm-network.selfLink)',
          'config': '$(ref.swarm-config.selfLink)'
      }
  }, {
      'name': 'workers',
      'type': 'workers.py',
      'properties': {
          'zone': zone,
          'template': '$(ref.worker.name)',
          'size': workerCount
      }
  }, {
      'name': 'swarm-network',
      'type': 'network.py'
  }, {
      'name': 'firewall-rules',
      'type': 'firewall.py',
      'properties': {
          'network': 'swarm-network'
      }
  }, {
      'name': 'swarm-config',
      'type': 'config.py'
  }]
  return {'resources': resources}

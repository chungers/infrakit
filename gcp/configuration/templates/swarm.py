# Copyright 2016 Docker Inc. All rights reserved.

"""Creates the Swarm."""

def GenerateConfig(context):
  zone = context.properties['zone']
  machineType = context.properties['machineType']
  size = context.properties['size']

  resources = [{
      'name': 'docker',
      'type': 'templates/disk-image.py'
  }, {
      'name': 'manager',
      'type': 'templates/manager.py',
      'properties': {
          'zone': zone,
          'machineType': machineType,
          'image': '$(ref.docker.selfLink)',
          'network': 'swarm-network'
      }
  }, {
      'name': 'worker',
      'type': 'templates/worker.py',
      'properties': {
          'zone': zone,
          'machineType': machineType,
          'image': '$(ref.docker.selfLink)',
          'network': 'swarm-network',
          'managerIP': '$(ref.manager.internalIP)'
      }
  }, {
      'name': 'workers',
      'type': 'templates/workers.py',
      'properties': {
          'zone': zone,
          'template': '$(ref.worker.name)',
          'size': size
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

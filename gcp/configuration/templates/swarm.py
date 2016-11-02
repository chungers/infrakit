# Copyright 2016 Docker Inc. All rights reserved.

"""Creates the Swarm."""

NETWORK_NAME = 'swarm-network'

def GenerateConfig(context):
  resources = [{
      'name': 'manager',
      'type': 'templates/manager.py',
      'properties': {
          'machineType': context.properties['machineType'],
          'zone': context.properties['zone'],
          'network': NETWORK_NAME
      }
  }, {
      'name': 'worker',
      'type': 'templates/worker.py',
      'properties': {
          'machineType': context.properties['machineType'],
          'zone': context.properties['zone'],
          'network': NETWORK_NAME,
          'managerIP': '$(ref.manager.internalIP)'
      }
  }, {
      'name': 'workers',
      'type': 'templates/workers.py',
      'properties': {
          'zone': context.properties['zone'],
          'template': '$(ref.worker.name)',
          'size': context.properties['size']
      }
  }, {
      'name': NETWORK_NAME,
      'type': 'templates/network.py'
  }, {
      'name': 'firewall-rules',
      'type': 'templates/firewall.py',
      'properties': {
          'network': NETWORK_NAME
      }
  }]
  return {'resources': resources}

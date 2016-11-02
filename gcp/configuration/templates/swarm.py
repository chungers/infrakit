# Copyright 2016 Docker Inc. All rights reserved.

"""Creates the Swarm."""

ZONE = 'europe-west1-d'
MACHINE_TYPE = 'g1-small'
NETWORK_NAME = 'swarm-network'

def GenerateConfig(context):
  resources = [{
      'name': 'manager',
      'type': 'templates/manager.py',
      'properties': {
          'machineType': MACHINE_TYPE,
          'zone': ZONE,
          'network': NETWORK_NAME
      }
  }, {
      'name': 'worker',
      'type': 'templates/worker.py',
      'properties': {
          'machineType': MACHINE_TYPE,
          'zone': ZONE,
          'network': NETWORK_NAME,
          'managerIP': '$(ref.manager.internalIP)'
      }
  }, {
      'name': 'workers',
      'type': 'templates/workers.py',
      'properties': {
          'zone': ZONE,
          'template': '$(ref.worker.name)'
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

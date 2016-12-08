# Copyright 2016 Docker Inc. All rights reserved.

"""Creates the Swarm."""

def GenerateConfig(context):
  region = context.properties['region']
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
      'name': 'docker-ip',
      'type': 'ip.py',
      'properties': {
          'region': region
      }
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
          'size': managerCount,
          'pool': '$(ref.docker-pool.selfLink)'
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
          'size': workerCount,
          'pool': '$(ref.docker-pool.selfLink)'
      }
  }, {
      'name': 'docker-pool',
      'type': 'pool.py',
      'properties': {
          'region': region
      }
  }, {
      'name': 'forwarding',
      'type': 'forwarding.py',
      'properties': {
          'region': region,
          'pool': '$(ref.docker-pool.selfLink)',
          'ip': '$(ref.docker-ip.address)'
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

  outputs = [{
      'name': 'externalIp',
      'value': '$(ref.docker-ip.address)'
  }]

  return {'resources': resources, 'outputs': outputs}

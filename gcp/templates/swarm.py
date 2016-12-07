# Copyright 2016 Docker Inc. All rights reserved.

"""Configure the project to host a Swarm."""

def GenerateConfig(context):
  region = context.properties['region']
  zone = context.properties['zone']
  managerMachineType = context.properties['managerMachineType']
  managerCount = context.properties['managerCount']
  workerCount = context.properties['workerCount']

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
      'name': 'docker-pool',
      'type': 'pool.py',
      'properties': {
          'region': region,
          'firstInstance': '$(ref.manager-1.selfLink)'
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
      'name': 'manager-1',
      'type': 'leader.py',
      'properties': {
          'zone': zone,
          'machineType': managerMachineType,
          'image': '$(ref.docker.selfLink)',
          'network': '$(ref.swarm-network.selfLink)',
          'config': '$(ref.swarm-config.selfLink)',
          'managerCount': managerCount,
          'workerCount': workerCount
      }
  }, {
      'name': 'swarm-config',
      'type': 'config.py'
  }]

  outputs = [{
      'name': 'externalIp',
      'value': '$(ref.docker-ip.address)'
  },{
      'name': 'leaderIp',
      'value': '$(ref.manager-1.networkInterfaces[0].accessConfigs[0].natIP)'
  },{
      'name': 'zone',
      'value': zone
  }]

  return {'resources': resources, 'outputs': outputs}

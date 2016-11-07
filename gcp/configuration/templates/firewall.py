# Copyright 2016 Docker Inc. All rights reserved.

"""Firewall rules for http/https/ssh and internal swarm communication."""

def GenerateConfig(context):
  network = '$(ref.' + context.properties['network'] + '.selfLink)'

  resources = [{
      'name': 'ssh',
      'type': 'compute.v1.firewall',
      'properties': {
          'network': network,
          'sourceRanges': ['0.0.0.0/0'],
          'allowed': [{
              'IPProtocol': 'TCP',
              'ports': [22]
          }]
      }
  },{
      'name': 'http',
      'type': 'compute.v1.firewall',
      'properties': {
          'network': network,
          'sourceRanges': ['0.0.0.0/0'],
          'allowed': [{
              'IPProtocol': 'TCP',
              'ports': [80]
          },{
              'IPProtocol': 'TCP',
              'ports': [443]
          }]
      }
  },{
      'name': 'internal',
      'type': 'compute.v1.firewall',
      'properties': {
          'network': network,
          'sourceTags': ['swarm'],
          'allowed': [{
              'IPProtocol': 'TCP'
          }]
      }
  }]

  return {'resources': resources}

# Copyright 2016 Docker Inc. All rights reserved.

"""Firewall rules for http/https/ssh and internal swarm communication."""

def GenerateConfig(context):
  network = '$(ref.' + context.properties['network'] + '.selfLink)'

  resources = [{
      'name': 'allow-ssh',
      'type': 'compute.v1.firewall',
      'properties': {
          'network': network,
          'sourceRanges': ['0.0.0.0/0'],
          'allowed': [{
              'IPProtocol': 'tcp',
              'ports': ['22']
          }]
      }
  },{
      'name': 'allow-http',
      'type': 'compute.v1.firewall',
      'properties': {
          'network': network,
          'sourceRanges': ['0.0.0.0/0'],
          'allowed': [{
              'IPProtocol': 'tcp',
              'ports': ['80', '443']
          }]
      }
  },{
      'name': 'allow-internal',
      'type': 'compute.v1.firewall',
      'properties': {
          'network': network,
          'sourceRanges': ['10.128.0.0/9'],
          'allowed': [{
              'IPProtocol': 'tcp',
              "ports": ['0-65535']
          },{
              'IPProtocol': 'udp',
              "ports": ['0-65535']
          }]
      }
  }]

  return {'resources': resources}

# Copyright 2016 Docker Inc. All rights reserved.

"""Swarm's network."""

def GenerateConfig(context):
  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.network',
      'properties': {
          'autoCreateSubnetworks': True
      }
  }]
  return {'resources': resources}

# Copyright 2016 Docker Inc. All rights reserved.

"""Swarm External IP."""

def GenerateConfig(context):
  region = context.properties['region']

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.address',
      'properties': {
          'region': region
      }
  }]

  return {'resources': resources}

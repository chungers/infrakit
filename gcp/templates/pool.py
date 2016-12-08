# Copyright 2016 Docker Inc. All rights reserved.

"""Swarm External IP."""

def GenerateConfig(context):
  region = context.properties['region']

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.targetPool',
      'properties': {
          'region': region,
          'healthChecks': [],
          'sessionAffinity': 'NONE'
      }
  }]

  return {'resources': resources}

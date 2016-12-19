# Copyright 2016 Docker Inc. All rights reserved.

"""Swarm External IP."""

def GenerateConfig(context):
  region = context.properties['region']
  firstInstance = context.properties['firstInstance']

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.targetPool',
      'properties': {
          'region': region,
          'healthChecks': [],
          'sessionAffinity': 'NONE',
          'instances': [firstInstance]
      }
  }]

  return {'resources': resources}

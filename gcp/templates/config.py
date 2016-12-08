# Copyright 2016 Docker Inc. All rights reserved.

"""Swarm Runtime Configuration."""

def GenerateConfig(context):
  resources = [{
      'name': context.env['name'],
      'type': 'runtimeconfig.v1beta1.config',
      'properties': {
          'config': context.env['name']
      }
  }]

  return {'resources': resources}

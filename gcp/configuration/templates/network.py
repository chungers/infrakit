# Copyright 2016 Docker Inc. All rights reserved.

"""Creates the network."""

def GenerateConfig(context):
  """Creates the network."""

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.network',
      'properties': {
          'IPv4Range': '10.0.0.1/16'
      }
  }]
  return {'resources': resources}

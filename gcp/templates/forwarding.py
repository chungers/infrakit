# Copyright 2016 Docker Inc. All rights reserved.

"""Forward external traffic to the nodes."""

def GenerateConfig(context):
  region = context.properties['region']
  pool = context.properties['pool']
  ip = context.properties['ip']

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.forwardingRule',
      'properties': {
          'region': region,
          'IPProtocol': 'TCP',
          'portRange': '80-65535',
          'IPAddress': ip,
          'target': pool,
      }
  }]

  return {'resources': resources}

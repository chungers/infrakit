# Copyright 2016 Docker Inc. All rights reserved.

"""Manager Instance Group."""

def GenerateConfig(context):
  project = context.env['project']
  zone = context.properties['zone']
  size = context.properties['size']
  template = context.properties['template']

  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.instanceGroupManager',
      'properties': {
          'zone': zone,
          'instanceTemplate': '/'.join(['projects', project,
                                       'global',
                                       'instanceTemplates', template]),
          'baseInstanceName': context.env['name'],
          'targetSize': size,
          'autoHealingPolicies': [{
              'initialDelaySec': 300
          }]
      }
  }]
  return {'resources': resources}

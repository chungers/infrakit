# Copyright 2016 Docker Inc. All rights reserved.

"""Worker Instance Group Manager."""

def GenerateConfig(context):
  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.instanceGroupManager',
      'properties': {
          'zone': context.properties['zone'],
          'instanceTemplate': '/'.join(['projects', context.env['project'],
                                       'global', 'instanceTemplates',
                                       context.properties['template']]),
          'baseInstanceName': context.env['name'],
          'targetSize': context.properties['size'],
          'autoHealingPolicies': [{
              'initialDelaySec': 300
          }]
      }
  }]
  return {'resources': resources}

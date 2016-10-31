# Copyright 2016 Docker Inc. All rights reserved.

"""Creates the virtual machine."""

COMPUTE_URL_BASE = 'https://www.googleapis.com/compute/v1'
PROJECT = 'code-story-blog'
ZONE = 'europe-west1-d'
MACHINE_TYPE = 'g1-small'

def GenerateConfig(context):
  """Creates a virtual machine."""

  resources = [{
      'name': context.properties['name'],
      'type': 'compute.v1.instance',
      'properties': {
          'zone': ZONE,
          'machineType': '/'.join([COMPUTE_URL_BASE, 'projects', PROJECT,
                                  'zones', ZONE,
                                  'machineTypes', MACHINE_TYPE]),
          'disks': [{
              'deviceName': 'boot',
              'type': 'PERSISTENT',
              'boot': True,
              'autoDelete': True,
              'initializeParams': {
                  'sourceImage': '/'.join([COMPUTE_URL_BASE, 'projects',
                                          'debian-cloud/global',
                                          'images/family/debian-8'])
              }
          }],
          'networkInterfaces': [{
              'network': '$(ref.docker-network.selfLink)',
              'accessConfigs': [{
                  'name': 'External NAT',
                  'type': 'ONE_TO_ONE_NAT'
              }]
          }]
      }
  }]
  return {'resources': resources}

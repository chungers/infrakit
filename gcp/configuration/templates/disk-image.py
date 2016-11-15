# Copyright 2016 Docker Inc. All rights reserved.

"""Moby disk image used for all the instances."""

def GenerateConfig(context):
  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.image',
      'properties': {
          'family': 'docker',
          'rawDisk': {
              'source': 'https://storage.cloud.google.com/docker-image/docker.image.tar.gz'
          }
      }
  }]

  return {'resources': resources}

# Copyright 2016 Docker Inc. All rights reserved.

"""Initial leader."""

def ManagerIds(managerCount):
  return '[%s]' % ','.join('"manager-%d"' % i for i in range(1,int(managerCount)+1))

def SplitLines(script):
  return ','.join('"%s"' % (line.replace('"', '\\"')) for line in script.split('\n'))

def InfrakitJson(context, type, allocationType, allocation, machineType, script):
  return context.imports['infrakit.json'] % (type + 's', allocationType, allocation, machineType, context.properties['network'], type, 10, context.properties['diskImage'], context.env['deployment'] + '-target-pool', SplitLines(script), type)

def ManagerJson(context, count, machineType, script):
  return InfrakitJson(context, 'manager', 'LogicalIDS', ManagerIds(count), machineType, script)

def WorkerJson(context, count, machineType, script):
  return InfrakitJson(context, 'worker', 'Size', count, machineType, script)

def GenerateConfig(context):
  resources = [{
      'name': context.env['name'],
      'type': 'compute.v1.instance',
      'properties': {
          'zone': context.properties['zone'],
          'machineType': 'zones/' + context.properties['zone'] + '/machineTypes/' + context.properties['managerMachineType'],
          'tags': {
              'items': ['swarm']
          },
          'disks': [{
              'deviceName': 'boot',
              'type': 'PERSISTENT',
              'boot': True,
              'autoDelete': True,
              'initializeParams': {
                  'sourceImage': context.properties['diskImage']
              }
          }],
          'networkInterfaces': [{
              'network': context.properties['network'],
              'accessConfigs': [{
                  'name': 'External NAT',
                  'type': 'ONE_TO_ONE_NAT'
              }]
          }],
          'metadata': {
              'items': [{
                  'key': 'startup-script',
                  'value': context.imports['manager-startup.sh'] + context.imports['leader-startup.sh']
              }, {
                  'key': 'managersJson',
                  'value': ManagerJson(context, context.properties['managerCount'], context.properties['managerMachineType'], context.imports['manager-startup.sh'])
              }, {
                  'key': 'workersJson',
                  'value': WorkerJson(context, context.properties['workerCount'], context.properties['workerMachineType'], "")
              }]
          },
          'scheduling': {
              'preemptible': False,
              'onHostMaintenance': 'TERMINATE',
              'automaticRestart': False
          },
          'serviceAccounts': [{
              'scopes': [
                  'https://www.googleapis.com/auth/cloudruntimeconfig',
                  'https://www.googleapis.com/auth/logging.write',
                  'https://www.googleapis.com/auth/compute'
              ]
          }]
      }
  }]
  return {'resources': resources}

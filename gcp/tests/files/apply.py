import os, sys
from jinja2 import Environment, FileSystemLoader

env = Environment(loader = FileSystemLoader('/templates'))
template = env.get_template(sys.argv[1])

print template.render({
    'properties': {
        'version': 'latest',
        'editions_version': 'latest',
        'managerCount': 3,
        'workerCount': 1,
        'zone': 'europe-west1-d',
        'managerMachineType': 'g1-small',
        'managerDiskType': 'pd-standard',
        'managerBootDiskSize': 20,
        'managerDiskSize': 100,
        'workerMachineType': 'g1-small',
        'workerDiskType': 'pd-standard',
        'workerDiskSize': 100,
        'preemptible': False,
        'enableSystemPrune': True,
        'remoteSshOrigin': '0.0.0.0/0'
    },
    'env': {
        'project': 'test-project',
        'deployment': 'docker'
    },
    'type':'manager'
})

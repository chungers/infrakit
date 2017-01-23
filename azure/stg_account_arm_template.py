#!/usr/bin/python

"""
Script to read 'editions.json' ("production" template designed to launch VMs
from the Azure Marketplace) and generate a temporary ARM template used to
launch a VM from the given storage account instead.
"""

import sys, json

if len(sys.argv) != 3:
	sys.stderr.write('Usage: python stg_account_arm_template.py [VHD_URL] [OUTPUT_FILE]\n')
	sys.stderr.write('Must pass VHD in storage account.\n')
	exit(1)

storageaccount = 'dockereditions'
prod_template = 'editions.json'
desired_vhd = sys.argv[1]
dev_template = sys.argv[2]

def launch_vm_from_stg_account(resource):
	if resource['type'] == 'Microsoft.Compute/virtualMachines':
		resource.pop('plan', None)
		resource['properties']['storageProfile'].pop('imageReference', None)
		resource['properties']['storageProfile']['osDisk']['image'] = {'uri': desired_vhd}
		resource['properties']['storageProfile']['osDisk']['osType'] = 'Linux'

		vhd_directive = "[concat('{}/', uniqueString(resourceGroup().id), '-manager.vhd')]".format('/'.join(desired_vhd.split('/')[:3] + ['vhds']))
		resource['properties']['storageProfile']['osDisk']['vhd'] = {'uri': vhd_directive}

	if resource['type'] == 'Microsoft.Compute/virtualMachineScaleSets':
		resource.pop('plan', None)
		resource['properties']['virtualMachineProfile']['storageProfile'].pop('imageReference', None)
		resource['properties']['virtualMachineProfile']['storageProfile']['osDisk']['image'] = {'uri': desired_vhd}
		resource['properties']['virtualMachineProfile']['storageProfile']['osDisk']['osType'] = 'Linux'
		resource['properties']['virtualMachineProfile']['storageProfile']['osDisk'].pop('vhdContainers', None)

	return resource

def gen_template(in_template, out_template):
	with open(in_template, 'r') as arm_template:
		arm_json = json.loads(arm_template.read())
		new_resouces = [launch_vm_from_stg_account(resource) for resource in arm_json['resources']]
		arm_json['resources'] = new_resouces
		with open(out_template, 'w') as generated_arm_template:
			generated_arm_template.write(json.dumps(arm_json, indent=4, sort_keys=True))
		print('Generated {} for rapid development').format(out_template)


if __name__ == '__main__':
	gen_template(prod_template, dev_template)
	print('Now you\'re thinking without portals.')

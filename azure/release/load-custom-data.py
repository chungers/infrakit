#!/usr/bin/env python

# Parses the custom data file for Manager and Worker and injects them into the Azure template

from collections import OrderedDict
import os
import re
import json



def buildCustomData(data_file):
	customData = []
	with open(data_file) as f:
		for line in f:
			m = re.match(r'(.*?)((?:parameters|variables)\([^\)]*\))(.*$)', line)
			if m:
				customData += ['\'' + m.group(1) + '\'',
							m.group(2),
							'\'' + m.group(3) + '\'',
							'\'\n\'']
			else:
				customData += ['\'' + line + '\'']
	return customData

manager_data = buildCustomData('custom-data_manager.sh')
worker_data = buildCustomData('custom-data_worker.sh')


with open('editions.template.json') as f:
	templ = json.load(f, object_pairs_hook=OrderedDict)
	templ['variables']['customDataManager'] = '[concat(' + ', '.join(
		manager_data) + ')]'
	templ['variables']['customDataWorker'] = '[concat(' + ', '.join(
		worker_data) + ')]'

# os.rename('editions.json', 'editions.json.old')

with open('editions.json', 'w') as f:
	f.write(json.dumps(templ, indent=2).replace(' \n', '\n') + '\n')

#!/usr/bin/env python

import os
import sys
import json
import difflib


DEBUG = os.environ.get('DEBUG')

def main(argv):
	if len(argv) == 0:
		print 'sanity.py <inputfile.json>'
		sys.exit(2)
	jsonInput = argv[0]
	# Read the input file and split into list of strings
	with open(jsonInput) as myfile:
		dataList = myfile.read().splitlines()
	# Read the input file, parse it via Python and create list of strings
	with open(jsonInput) as data_file:
		dataJson = json.dumps(json.load(data_file), sort_keys=True, indent=4).splitlines()
	diff = []

	for line in difflib.context_diff(dataList, dataJson):
		diff.append(line)
	if len(diff) > 0:
		if DEBUG == "1":
			print("\n".join(diff))
		exit(1)
	else:
		exit(0)

if __name__ == '__main__':
	main(sys.argv[1:])
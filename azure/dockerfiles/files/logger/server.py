#!/usr/bin/env python

## Tiny Syslog Server in Python.
##
## This is a tiny syslog server that is able to receive UDP based syslog
## entries on a specified port and save them to a file.
## That's it... it does nothing else...
## There are a few configuration parameters.

LOG_PATH = '/log/'
HOST, PORT = "0.0.0.0", 514

#
# NO USER SERVICEABLE PARTS BELOW HERE...
#


# Azure log tactic:
# 1. Use hostname to create folder in files
# 2. Use container ID to create file
# 3. Push data to file

import logging
import SocketServer
import os
import argparse
import sys
from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.storage import StorageManagementClient
from azure.storage.file import (FileService, ContentSettings)
from pyparsing import Word, alphas, Suppress, Combine, nums, string, Optional, Regex
from time import strftime

logging.basicConfig(level=logging.INFO, format='%(message)s', datefmt='', filemode='a')

class Parser(object):
	def __init__(self):
		ints = Word(nums)
		# priority
		priority = Suppress("<") + ints + Suppress(">")
		# timestamp
		month = Word(string.uppercase, string.lowercase, exact=3)
		day   = ints
		hour  = Combine(ints + ":" + ints + ":" + ints)
		timestamp = month + day + hour
		# container name
		containername = Word(alphas + nums + "_" + "-" + ".") + Suppress("/")
		# containerid
		containerid = Word(alphas + nums) + Optional(Suppress("[") + ints + Suppress("]")) + Suppress(":")
		# message
		message = Regex(".*")
		# pattern build
		self.__pattern = priority + timestamp + containername + containerid + message
		
	def parse(self, line):
		parsed = self.__pattern.parseString(line)
		payload              = {}
		payload["priority"]  = parsed[0]
		payload["timestamp"] = strftime("%Y-%m-%d %H:%M:%S")
		payload["containername"]   = parsed[4]
		payload["containerid"]   = parsed[5]
		payload["pid"]       = parsed[6]
		payload["message"]   = parsed[7]
		
		return payload

class AzureLog(object):
	def __init__(self):
		self.__key = self.get_storage_key()
		self.file_service = self.create_file_service(self.__key)
		
	def write(self, container_log, message, level=logging.INFO):
		global LOG_PATH
		l = logging.getLogger(container_log)
		if not len(l.handlers):
			log_file = r"{0}{1}.log".format(LOG_PATH, container_log)
			fileHandler = logging.FileHandler(log_file, mode='a')
			l.addHandler(fileHandler)
		l.info(message)
		return l

	def get_storage_key(self):
		sub_id = os.environ['ACCOUNT_ID']
		cred = ServicePrincipalCredentials(
				client_id=os.environ['APP_ID'],
				secret=os.environ['APP_SECRET'],
				tenant=os.environ['TENANT_ID']
		)
		storage_client = StorageManagementClient(cred, sub_id)
		rg_name = os.environ['GROUP_NAME']
		sa_name = os.environ['SWARM_LOGS_STORAGE_ACCOUNT']
		storage_keys = storage_client.storage_accounts.list_keys(rg_name, sa_name)
		storage_keys = {v.key_name: v.value for v in storage_keys.keys}
		return storage_keys['key1']

	def create_file_service(self, storage_key):
		file_service = FileService(account_name=os.environ['SWARM_LOGS_STORAGE_ACCOUNT'], account_key=storage_key)
		file_service.create_share(os.environ['SWARM_FILE_SHARE'])
		return file_service

	def upload(self, container_log):
		log_file = r"{0}{1}.log".format(LOG_PATH, container_log)
		self.file_service.create_file_from_path(
			os.environ['SWARM_FILE_SHARE'], # File share name
			None, # We want to create this blob in the root directory, so we specify None for the directory_name
			container_log, # file name on share
			log_file, # the file to upload
			content_settings=ContentSettings(content_type='text/plain')) # content type
	
	# @TODO implement cleanup for container


class SyslogUDPHandler(SocketServer.BaseRequestHandler):
	
	def handle(self):
		global azure_log
		data = bytes.decode(self.request[0].strip())
		socket = self.request[1]
		parser = Parser()
		fields = parser.parse(str(data))
		print "parsed:", fields
		print( "%s : " % self.client_address[0], str(data))
		container_log = "{0}-{1}".format(fields["containername"], fields["containerid"])
		azure_log.write(container_log, fields["message"])
		azure_log.upload(container_log)
		# logging.info(fields)

if __name__ == "__main__":
	try:
		azure_log = AzureLog()
		server = SocketServer.UDPServer((HOST,PORT), SyslogUDPHandler)
		server.serve_forever(poll_interval=0.5)
	except (IOError, SystemExit):
		raise
	except KeyboardInterrupt:
		print ("Crtl+C Pressed. Shutting down.")

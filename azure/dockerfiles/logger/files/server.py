#!/usr/bin/env python

## Syslog Server that streams docker container logs over SMB to Azure file storage
##
## It receive UDP based syslog entries on a specified port and streams them 
## to files on a SMB share hosted on Azure File Storage.

LOG_MNT_PATH = '/logmnt/'
LOGS_SHARE_NAME = 'docker4azurelogs'
HOST, PORT = "0.0.0.0", 514

CIFS_OPTION_VERSION  = "vers=2.1"
CIFS_OPTION_FILE_MODE = "file_mode=0777"
CIFS_OPTION_DIR_MODE = "dir_mode=0777"
CIFS_OPTION_UID = "uid=0"
CIFS_OPTION_GID = "gid=0"

BUFFER_SZ_MAX = 1024*4     # 4 KB
BUFFER_DURATION_MAX = 30   # 30 seconds
CLEANUP_INTERVAL = 60*10   # 10 mins
IDLE_DURATION = 60*30      # 30 mins

# Azure log tactic:
# 1. Mount file storage share in logging storage account over SMB
# 2. Use container ID to create log files in mounted share
# 3. Flush log files when buffer size or buffering duration exceeds thresholds

import logging
import SocketServer
import os
import argparse
import sys
import datetime
import subprocess
from azure.common.credentials import ServicePrincipalCredentials
from azure.mgmt.storage import StorageManagementClient
from azure.storage.file import (FileService, ContentSettings)
from pyparsing import Word, alphas, Suppress, Combine, nums, string, Optional, Regex
from time import strftime
from azendpt import *

RESOURCE_MANAGER_ENDPOINT = os.getenv('RESOURCE_MANAGER_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['RESOURCE_MANAGER_ENDPOINT'])
ACTIVE_DIRECTORY_ENDPOINT = os.getenv('ACTIVE_DIRECTORY_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['ACTIVE_DIRECTORY_ENDPOINT'])
STORAGE_ENDPOINT = os.getenv('STORAGE_ENDPOINT', AZURE_PLATFORMS[AZURE_DEFAULT_ENV]['STORAGE_ENDPOINT'])

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
        self.file_service = self.create_file_service(self.__key, LOGS_SHARE_NAME)
        self.account_name = os.environ['SWARM_LOGS_STORAGE_ACCOUNT']
        self.mount_log_share(self.account_name, self.__key, LOGS_SHARE_NAME)

        self.log_file_handles = {}
        self.log_buffer_sizes = {}
        self.log_flush_timestamps = {}
        self.last_cleanup = datetime.datetime.utcnow()

    def mount_log_share(self, account_name, account_key, share_name):
        mount_uri = "//" + account_name + ".file." + STORAGE_ENDPOINT + "/" + share_name

        mount_opts = [
            CIFS_OPTION_VERSION,
            "username=" + self.account_name,
            "password=" + self.__key,
            CIFS_OPTION_FILE_MODE,
            CIFS_OPTION_DIR_MODE,
            CIFS_OPTION_UID,
            CIFS_OPTION_GID
        ]

        if not os.path.exists(LOG_MNT_PATH):
            os.makedirs(LOG_MNT_PATH)

        mount_cmd = "mount -t cifs " + mount_uri + " " + LOG_MNT_PATH + " -o " + ",".join(mount_opts)
        # raise on failure. If we can't mount the share, there's nothing we can do .. so die.
        subprocess.check_call(mount_cmd, shell=True)  


    def write(self, container_id, message):
        # called in the context of a UDP recv call. Avoid calling blocking/remote APIs!
        container_log_path = LOG_MNT_PATH + "/" + container_id + ".log"
        if not container_id in self.log_file_handles:
            self.log_file_handles[container_id] = open(container_log_path, 'a')
            self.log_buffer_sizes[container_id] = 0
            self.log_flush_timestamps[container_id] = datetime.datetime.utcnow()

        buffer_sz = self.log_buffer_sizes[container_id]
        buffer_duration = datetime.datetime.utcnow() - self.log_flush_timestamps[container_id]
        # reset the buffer and file handle to flush all contents at regular intervals
        
        if ((buffer_sz != 0 and buffer_sz + len(message) > BUFFER_SZ_MAX) or 
            (buffer_duration.seconds > BUFFER_DURATION_MAX)):
            self.log_file_handles[container_id].close()
            self.log_file_handles[container_id] = open(container_log_path, 'a')
            buffer_sz = 0
            self.log_flush_timestamps[container_id] = datetime.datetime.utcnow()

        try:
            self.log_file_handles[container_id].write(message + "\n")
        except IOError as e:
            print ("I/O error({0}): {1}".format(e.errno, e.strerror))
        self.log_buffer_sizes[container_id] = buffer_sz + len(message)

        #check if we need to garbage collect old container log entries
        if (datetime.datetime.utcnow() - self.last_cleanup).seconds > CLEANUP_INTERVAL:
            self.cleanup_file_entries(True)

        return

    def get_storage_key(self):
        sub_id = os.environ['ACCOUNT_ID']
        cred = ServicePrincipalCredentials(
                client_id=os.environ['APP_ID'],
                secret=os.environ['APP_SECRET'],
                tenant=os.environ['TENANT_ID'],
                resource=RESOURCE_MANAGER_ENDPOINT,
                auth_uri=ACTIVE_DIRECTORY_ENDPOINT
        )
        storage_client = StorageManagementClient(cred, sub_id, base_url=RESOURCE_MANAGER_ENDPOINT)
        rg_name = os.environ['GROUP_NAME']
        sa_name = os.environ['SWARM_LOGS_STORAGE_ACCOUNT']
        storage_keys = storage_client.storage_accounts.list_keys(rg_name, sa_name)
        storage_keys = {v.key_name: v.value for v in storage_keys.keys}
        return storage_keys['key1']

    def create_file_service(self, storage_key, share_name):
        file_service = FileService(account_name=os.environ['SWARM_LOGS_STORAGE_ACCOUNT'], account_key=storage_key, endpoint_suffix=STORAGE_ENDPOINT)
        file_service.create_share(share_name)
        return file_service     

    def cleanup_file_entries(self, check_time):
        for container_id in self.log_file_handles.keys():
            if check_time:
                if ((self.log_flush_timestamps[container_id] - datetime.datetime.utcnow()).seconds > IDLE_DURATION):
                    self.cleanup_file_entry(container_id)
            else:
                self.cleanup_file_entry(container_id)
        self.last_cleanup = datetime.datetime.utcnow()

    def cleanup_file_entry(self, container_id):
        self.log_file_handles[container_id].close()
        del self.log_file_handles[container_id]
        del self.log_buffer_sizes[container_id]
        del self.log_flush_timestamps[container_id]   
	# @TODO implement cleanup for container


class SyslogUDPHandler(SocketServer.BaseRequestHandler):
	
    def handle(self):
        global azure_log
        data = bytes.decode(self.request[0].strip())
        socket = self.request[1]
        parser = Parser()
        fields = parser.parse(str(data))
        container_id = "{0}-{1}".format(fields["containername"], fields["containerid"])
        azure_log.write(container_id, fields["message"])

if __name__ == "__main__":
    try:
        azure_log = AzureLog()
        server = SocketServer.UDPServer((HOST,PORT), SyslogUDPHandler)
        server.serve_forever(poll_interval=0.5)
    except (IOError, SystemExit):
        raise
    except KeyboardInterrupt:
        print ("Crtl+C Pressed. Shutting down.")

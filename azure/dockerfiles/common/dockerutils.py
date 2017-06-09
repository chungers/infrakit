#!/usr/bin/env python

import logging
import subprocess
from time import sleep

DOCKER_NODE_STATUS_READY = "ready"
DOCKER_NODE_STATUS_DOWN = "down"

def remove_overprovisioned_nodes(client, node_hostname, logger):
    logger.info("Check for any overprovisioned unreachable nodes and remove them ...")
    retry = True
    keys_missing = False
    while retry:
        for node in client.nodes():
            try:
                node_id = node['ID']
                if (node['Description']['Hostname'] == node_hostname) and (node['Status']['State'] == DOCKER_NODE_STATUS_DOWN):
                    # the detection and removal here depends on metadata server denying access to tokens 
                    # and thus serializing swarm joins by multiple overprovisioned nodes during swarm join
                    logger.info("Detected down node with hostname: {}".format(node_hostname))
                    cmdout = subprocess.check_output(["docker", "node", "ls"])
                    logger.info("docker node ls output: {}".format(cmdout))
                    logger.info("Remove overprovisioned down node with hostname: {}".format(node_hostname))
                    if node['Spec']['Role'] == 'manager':
                        cmdout = subprocess.check_output(["docker", "node", "demote", node_id])
                        logger.info("Demote manager node before removal: {}".format(cmdout))
                        sleep(10)
                    cmdout = subprocess.check_output(["docker", "node", "rm", "--force", node_id])
                    logger.info("docker node rm output: {}".format(cmdout))
                    cmdout = subprocess.check_output(["docker", "node", "ls"])
                    logger.info("docker node ls output: {}".format(cmdout))
            except KeyError:
                # When a member is joining, sometimes things
                # are a bit unstable and keys are missing. So retry.
                logger.info("Description/Hostname not found. Retrying ..")
                keys_missing = True
                break

        if keys_missing:
            sleep(10)
            continue
        else:
            return

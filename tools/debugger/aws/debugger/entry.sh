#!/usr/bin/env bash

DEBUG_VERSION="0.1"

function h1 () {
    echo " ======== " "$@" " ========"
    echo ""
}

function h2 () {
    echo " ------ " "$@" " ----------"
    echo ""
}

echo "##==================== Start Debugger =======================##"
echo "##==================== Version: $DEBUG_VERSION =======================##"
h2 `date`

h1 "Swarm Nodes"
docker node ls

NODES=$(docker node inspect $(docker node ls -q) | jq -r '.[] | .Description.Hostname')

for I in $NODES; do
    h1 "Start $I"
    docker -H $I:2375 pull docker4x/node-debug:$DEBUG_VERSION
    docker -H $I:2375 run -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/lib/docker/swarm:/var/lib/docker/swarm -v /var/log:/var/log docker4x/node-debug:$DEBUG_VERSION
    h1 "Finished $I"
done

h2 `date`
echo "##==================== Finished ====================##"

#!bin/bash

#Script to check the manager and worker counts are correct in deployment

#Values that we will check against
MANCOUNT=$1
WORKERCOUNT=$2

DOCKER_BIN=/usr/local/bin/docker

#Get the manager and worker count from docker node ls
WC=$( $DOCKER_BIN  node ls | grep "swarm-worker"| wc -l)
MC=$( $DOCKER_BIN  node ls | grep "swarm-manager" | wc -l)


$DOCKER_BIN node ls

#Check that the manager and worker counts both match up
if [[ $WORKERCOUNT -eq $WC && $MANCOUNT -eq $MC ]]; then
        exit 0
else
        exit 1
fi

#!bin/bash

#Script to check the manager and worker counts are correct in deployment

#Values that we will check against
MANCOUNT=$1
WORKERCOUNT=$2

EXPECTEDCOUNT=$(( $MANCOUNT + $WORKERCOUNT )) 
echo "Expected count is: $EXPECTEDCOUNT"

DOCKER_BIN=docker

#Get the manager and worker count from docker node ls
TOTALCOUNT=$( $DOCKER_BIN  node ls | wc -l )

#Minus one to get rid of the headers
TOTALCOUNT=$(( $TOTALCOUNT - 1 ))
echo "total count is: $TOTALCOUNT"

$DOCKER_BIN node ls

#Check that the manager and worker counts both match up
if [[ $EXPECTEDCOUNT -eq $TOTALCOUNT   ]]; then
        exit 0
else
        exit 1
fi

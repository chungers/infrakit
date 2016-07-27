#!/usr/bin/env bash

VERSION=0.1

# This is the script that will run the debugger container
LOGFILE="DEBUG_`date +'%Y-%m-%d_%H_%M_%s'`.log"
docker run \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/bin/docker \
docker4x/debugger:$VERSION 2>&1 > $LOGFILE

echo "Debug Collection Complete, your log is located in $LOGFILE"

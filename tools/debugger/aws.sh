#!/usr/bin/env bash

# This is the script that will run the debugger container
LOGFILE="DEBUG_`date +'%Y-%m-%d_%H_%M_%s'`.log"
docker run \
-v /var/run/docker.sock:/var/run/docker.sock \
-v /usr/bin/docker:/usr/docker/bin/docker \
docker4x/debugger:0.1 2>&1 > $LOGFILE

echo "Debug Collection Complete, your log is located in $LOGFILE"
echo "To download run the following command from you local machine."
PUBLIC_IP=$(wget -qO- http://169.254.169.254/latest/meta-data/public-ipv4)
echo "scp docker@$PUBLIC_IP:`pwd`/$LOGFILE /tmp"

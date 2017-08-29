#!/bin/bash

if [ "$#" -ne 1 ]; then
    echo "Target version tag missing"
    echo "USAGE: upgrade.sh YY.MM.S-latest"
    exit
fi

VERSION_TAG=$1

docker pull docker4x/upgrade-azure:$VERSION_TAG
if [ $? -ne 0 ]; then
    echo "Upgrade target $VERSION_TAG not found. Please confirm $VERSION_TAG has been released by Docker."
    exit
fi

echo "Execute upgrade container for $VERSION_TAG"

 # wrapper around the python upgrade subscriptions
 docker run \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /usr/bin/docker:/usr/bin/docker \
    -v /var/lib/waagent/CustomData:/var/lib/waagent/CustomData \
    -ti \
    docker4x/upgrade-azure-core:$VERSION_TAG

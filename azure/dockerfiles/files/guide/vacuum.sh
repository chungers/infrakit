#!/bin/bash

# need to query parameter using Azure APIs. Passing through customData will break upgrades today.
# TODO: pass this through customData once upgrade with customData is supported for Azure VMSS.
RUN_VACUUM=$(azparameters.py enableSystemPrune)
if [ $? -ne 0 ]; then
    exit 0
fi

if [[ "$RUN_VACUUM" != "yes" ]] ; then
    exit 0
fi

# sleep for a random amount of time, so that we don't run this at the same time on all nodes.
sleep $[ ( $RANDOM % 240 )  + 1 ]

docker system prune --force

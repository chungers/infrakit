#!/bin/bash

# wrapper around the python upgrade subscriptions
[ $# -ne 1 ] && { echo "Usage: $0 new-version"; exit 1; }
cont_id=$(docker ps -a | grep guide-azure | cut -d' ' -f1)
echo executing upgrade on $cont_id
docker exec -ti $cont_id /usr/bin/azupgrade.py $1

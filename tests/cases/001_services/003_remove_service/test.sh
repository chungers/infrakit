#!/bin/sh
# SUMMARY: Check to make sure a service can be removed properly
# LABELS:

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

NAME=test

# Create then remove the service
docker service create --detach=false --name "${NAME}" nginx
docker service remove "${NAME}"

# Check that there are no services running
docker service ls | grep "${NAME}" | assert_empty

exit 0
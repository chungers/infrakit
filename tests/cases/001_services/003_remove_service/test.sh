#!/bin/sh


# Check to make sure a service can be removed properly

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

NAME=test

# Create then remove the service
docker service create --detach=false --name "${NAME}" nginx
docker service remove "${NAME}"

# Check that there are no services running
docker service ls | tail +2 | assert_empty

exit 0
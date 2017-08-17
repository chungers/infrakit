#!/bin/sh

# Create a basic service with only 1 replica
set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

NAME=test

# remove the service created when script exits
clean_up() {
    docker service remove "${NAME}"
}
trap clean_up EXIT

# Create the service
# --detach=false stops commands from running until the service is created
docker service create --detach=false --name "${NAME}" nginx

# Make sure the service was created and is running
# awk '{ print $6 }' prints the CURRENT STATE of a service process
docker service ps "${NAME}" | grep "${NAME}" | awk '{ print $6 }' | assert_contains Running

exit 0
#!/bin/sh
# SUMMARY: Check that the correct number of replicas are created
# LABELS:

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

NAME=test

# Remove service on exit of the script
clean_up() {
    docker service remove "${NAME}"
}
trap clean_up EXIT

# Create a service with multiple replicas
docker service create --detach=false --replicas 2 --name "${NAME}" nginx

# Check the number of replicas
# check_replicas can be found in editions/tests/cases/_lib/lib.sh
NUM_REPLICAS=$(check_replicas "${NAME}" 2 5)
[ "${NUM_REPLICAS}" = 2 ]

exit 0
#!/bin/sh

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

NAME=test

clean_up() {
    docker service remove "${NAME}"
}

trap clean_up EXIT

docker service create --name "${NAME}" nginx
NUM_REPLICAS=$(check_replicas "${NAME}" 1 5)

[ "${NUM_REPLICAS}" = 1 ]
docker service ls | assert_contains "${NAME}"

exit 0
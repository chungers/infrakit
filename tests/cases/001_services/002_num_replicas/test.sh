#!/bin/sh

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

NAME=test

clean_up() {
    docker service remove "${NAME}"
}

trap clean_up EXIT

docker service create --replicas 2 --name "${NAME}" nginx

NUM_REPLICAS=$(check_replicas "${NAME}" 2 5)
[ "${NUM_REPLICAS}" = 2 ]

exit 0
#!/bin/sh

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

DRAIN_NODE=$(docker node ls | tail -1 | cut -d " " -f 1)

cleanup() {
    docker node update --availability active "${DRAIN_NODE}"
}

trap cleanup EXIT

docker node update --availability drain "${DRAIN_NODE}"
docker node inspect --format "{{.Spec.Availability}}" "${DRAIN_NODE}" | assert_contains "drain"

exit 0
#!/bin/sh

set -e

echo "${RUN_VACUUM}" | grep -iqF true || exit 0

DELAY=$(($RANDOM % 240))
echo Sleep for ${DELAY}s, so that we don\'t run this at the same time on all nodes.
sleep ${DELAY}

docker system prune --force

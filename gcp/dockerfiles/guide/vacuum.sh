#!/bin/sh

[ "${RUN_VACUUM}" == "yes" ] || exit 0

DELAY=$(($RANDOM % 240))
echo Sleep for ${DELAY}s, so that we don\'t run this at the same time on all nodes.
[ "${SLEEP}" == "no" ] || sleep $(($RANDOM % 240))

docker system prune --force

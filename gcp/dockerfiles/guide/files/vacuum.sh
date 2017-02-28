#!/bin/sh

shopt -s nocasematch
case "${RUN_VACUUM}" in
 "True") ;;
 *) exit 0;;
esac

DELAY=$(($RANDOM % 240))
echo Sleep for ${DELAY}s, so that we don\'t run this at the same time on all nodes.
sleep ${DELAY}

docker system prune --force

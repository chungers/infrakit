#!/bin/sh
# SUMMARY: Check that a node's availability can be switched between 'drain' and 'active'
# LABELS:

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

# Get the ID of the last node listed
NODE=$(docker node ls | tail -1 | cut -d " " -f 1)

docker node update --availability drain "${NODE}"
sleep 2s
docker node inspect --format "{{.Spec.Availability}}" "${NODE}" | assert_contains "drain" || { echo "Node not switched to drain" >&2; exit 1; }
docker node update --availability active "${NODE}"
sleep 2s
docker node inspect --format "{{.Spec.Availability}}" "${NODE}" | assert_contains "active" || { echo "Node not switched to active" >&2; exit 1; }

exit 0
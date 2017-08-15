#!/bin/sh

# Check that all nodes are active by default
set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

TOTAL_NODES=$(docker node ls | tail +2 | wc -l)
AVAILABLE_NODES=$(docker node ls | grep Active | wc -l)

[ "${TOTAL_NODES}" = "${AVAILABLE_NODES}" ]

exit 0
#!/bin/sh
# SUMMARY: Check the version of Docker running is the expected version
# LABELS:

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

CLIENT_VERSION=$(docker version --format '{{.Client.Version}}')

[ "${CLIENT_VERSION}" = "${EXPECTED_VERSION}" ]

exit 0
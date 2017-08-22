#!/bin/sh
# SUMMARY: Check the version of Docker running is the expected version
# LABELS:

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

if [ -z $EXPECTED_VERSION ];
then
    echoerr "No EXPECTED_VERSION set"; exit 1
fi

CLIENT_VERSION=$(docker version --format '{{.Client.Version}}')

[ "${CLIENT_VERSION}" = "${EXPECTED_VERSION}" ]

exit 0
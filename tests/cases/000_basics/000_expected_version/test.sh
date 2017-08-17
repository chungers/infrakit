#!/bin/sh

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"


if [ -z $(EXPECTED_VERSION) ];
then
    EXPECTED_VERSION='17.06.0-ce'
fi

CLIENT_VERSION=$(docker version --format '{{.Client.Version}}')

[ "${CLIENT_VERSION}" = "${EXPECTED_VERSION}" ]

exit 0
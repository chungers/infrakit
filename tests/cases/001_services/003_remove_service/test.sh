#!/bin/sh

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

NAME=test

docker service create --name "${NAME}" nginx
docker service remove "${NAME}"
docker service ls | tail +2 | assert_empty

exit 0
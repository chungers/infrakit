#!/bin/sh
# Summary: Create a container from a docker file, and ensure it runs as expected 
# Labels:

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"

IMAGE="test_container"

clean_up() {
docker rmi -f  $IMAGE 
}
trap clean_up EXIT

docker build . -f Dockerfile -t $IMAGE 
docker run --rm $IMAGE "hello" | assert_contains "hello"



exit 0

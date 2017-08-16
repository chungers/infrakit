#!/bin/sh
# Summary: Create a container from a docker file, and ensure it runs as expected 
# Labels:

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"

docker build . -f Dockerfile -t test_container
docker run --rm test_container "hello" | assert_contains "hello"
#docker run --rm test_container "hello" 

exit 0

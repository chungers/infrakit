#!/bin/sh
# SUMMARY: Creates a service without an exposed port. Exposes a port and verifies that the service has that port exposed
# LABELS:

REPS=3
NAME="ping3"

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"


clean_up() {
    docker service rm "${NAME}" || true
}
trap clean_up EXIT

# Deploy a service with no ports exposed
docker service create --detach=false --replicas $REPS --name $NAME alpine ping docker.com


# Check that the service is up and running (make this a library function later in some way or other)
ACTUAL=$(check_replicas $NAME $REPS 10)
echo "Acutal number of replicas $ACTUAL expected number of replicas $REPS"
assert_equals "Correct number of replicas" $ACTUAL $REPS 

# Check that service has no ports exposed
ACTUAL=$(docker service ls | grep $NAME | awk '{ print $6 }' | wc -w)
assert_equals "service has no ports exposed" $ACTUAL 0  

# Expose a port
EXPOSED_PORT=8000

docker service update --detach=false --publish-add $EXPOSED_PORT $NAME 


# Check that port is available
ACTUAL=$(docker service ls | grep $NAME | awk '{ print $6 }')
echo $ACTUAL | assert_contains $EXPOSED_PORT 
echo "Port was exposed"

# Delete the port

docker service update --detach=false --publish-rm $EXPOSED_PORT $NAME 


# Check that the port is not available
ACTUAL=$(docker service ls | grep $NAME | awk '{ print $6 }' | wc -w)
assert_equals "service has no ports exposed" $ACTUAL 0  


exit 0

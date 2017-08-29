#!/bin/sh
# Summary: Verify that a service doesn't deploy on unsatisfiable constraints 
# Labels:

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"

REPLICAS=3
SERVICE_NAME="Top"$REPLICAS
RETRIES=3

clean_up(){
  docker service rm $SERVICE_NAME || echo "true" 
}

trap clean_up EXIT

# Deploy a service with an unsatisfiable constraint 
# Make service in detached state because it should never have all replicas
docker service create  --detach=true --name $SERVICE_NAME --replicas $REPLICAS --constraint 'node.labels.dne == zzz' alpine top 

# Check the replicas
# Check for the actual number of replicas so the service waits for awhile
NUM_REPLICAS=$(check_replicas $SERVICE_NAME $REPLICAS $RETRIES) 

# Assert that the number from checking the replicas is 0 
assert_equals "Expect 0 replicas" 0 $NUM_REPLICAS 

exit 0

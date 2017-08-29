#!/bin/sh
# Summary: Verify that you can deploy a service based on node constraints based on role
# Labels:

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"

REPLICAS=3
SERVICE_NAME="Top"$REPLICAS

clean_up(){
  docker service rm $SERVICE_NAME || true
}

trap clean_up EXIT


# Deploy a service with a manager constraint 
docker service create  --detach=false --name $SERVICE_NAME --replicas $REPLICAS --constraint 'node.role == manager' alpine top 


# Verify that only the nodes that are managers have the service running
SERVICES_NODES=$(docker service ps $SERVICE_NAME -q | tail -n +2)
for line in $SERVICE_NODES
do
    ROLE=$(docker node inspect $line --format "{{.Spec.Role}}")
    assert_equals "Expecting manager role" $ROLE "manager"
done



exit 0

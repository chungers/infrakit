#!/bin/sh
# Summary: Basic secrets test
# Labels: 

# Description:
# 1. Create a secret
# 2. Check that the secret shows up with docker secret ls
# 3. Add a service with a secret
# 4. Check that the secret is accessible from one of the containers 

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"

SERVICE_NAME="nginx"
REPLICAS=1
SECRET_NAME="secret_test"
# Secret is troublesome when spaces are used
SECRET="shh_it's_a_secret"
KEY="current"
VALUE="true"

clean_up() {
     docker service rm $SERVICE_NAME 
     docker secret rm $SECRET_NAME
     docker node update --label-rm $KEY $CURRENT_ID
}

trap clean_up EXIT

# Create a secret
echo $SECRET | docker secret create $SECRET_NAME -
docker secret ls | assert_contains $SECRET_NAME

# Get current node id
CURRENT_ID=$(docker node inspect self --format "{{ .ID }}")

# Add a label to the current node 
docker node update --label-add $KEY=$VALUE $CURRENT_ID

# Add a service with a secret and force the service to be on the current node
docker service create --detach=false --constraint "node.labels.${KEY} == ${VALUE}" --name $SERVICE_NAME --secret $SECRET_NAME --replicas $REPLICAS $SERVICE_NAME 

# Verify that the services are up
NUM_REPLICAS=$(check_replicas $SERVICE_NAME 1 5)
assert_equals "Expect the service is up" $NUM_REPLICAS $REPLICAS


# Use docker exec to connect to service that has access to the secret and read the secret
CONTAINER_ID=$(docker ps --filter name=$SERVICE_NAME -q)
echo containerid is $CONTAINER_ID
echo $(docker exec $CONTAINER_ID cat /run/secrets/$SECRET_NAME)
VERIFY_SECRET=$( docker exec $CONTAINER_ID cat /run/secrets/$SECRET_NAME)

# Verify that the secret is what is expected
echo "The secret from the service is $VERIFY_SECRET"
assert_equals "Expect secret to be correct" $VERIFY_SECRET $SECRET

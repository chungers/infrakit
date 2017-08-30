#!/bin/sh
# Summary: Create a secret and verify it's what's expected try to access the secret in different ways, and make sure secret functionality works as expected
# Labels: 

# Description:
# 1. Create a secret 
# 2. Add a service with a secret
# 3. Go into a container and verify that the secret is what's expected
# 4. Commit a container, and verify that the secret is not accessible on the committed container
# 5. Try to remove the secret when the service is still up (should fail)
# 6. Do a service update to remove the secret
# 7. Check that the secret can not be accessed and that it has actually been removed

set -e
 . "${RT_PROJECT_ROOT}/_lib/lib.sh"


SERVICE_NAME="nginx"
REPLICAS=1
SECRET_NAME="secret_test"
SECRET="shh_it's_a_secret"
RETRIES=5
KEY="current"
VALUE="true"

clean_up() {
     docker service rm $SERVICE_NAME 
     docker secret rm $SECRET_NAME
     rm err
     docker node update --label-rm $KEY $CURRENT_ID
}

trap clean_up EXIT


# Create a secret
echo $SECRET | docker secret create $SECRET_NAME -
docker secret ls | assert_contains $SECRET_NAME

#Get the node id of the node we are ssh into
CURRENT_ID=$(docker node inspect self --format "{{ .ID }}")

# Add  a label to the current node
docker node update --label-add $KEY=$VALUE $CURRENT_ID

# Add a service with a secret
docker service create --detach=false --name $SERVICE_NAME --constraint "node.labels.${KEY} == ${VALUE}" --secret $SECRET_NAME --replicas $REPLICAS $SERVICE_NAME 

# Verify that the services are up
NUM_REPLICAS=$(check_replicas $SERVICE_NAME $REPLICAS $RETRIES)
assert_equals "Expect the service is up" $REPLICAS $NUM_REPLICAS 


# Use docker exec to connect to service that has access to the secret and read the secret
CONTAINER_ID=$(docker ps --filter name=$SERVICE_NAME -q)
VERIFY_SECRET=$( docker exec $CONTAINER_ID cat /run/secrets/$SECRET_NAME)

# Verify that the secret is what is expected
assert_equals "Expect secret to be correct" $VERIFY_SECRET $SECRET

# Verify that the secret is not available when committing the container
COMMITTED_CONTAINER=commited_$SERVICE_NAME
docker commit $CONTAINER_ID $COMMITTED_CONTAINER

VERIFY_SECRET=$(docker run --rm -t $COMMITTED_CONTAINER cat /run/secrets/$SECRET_NAME)
echo $VERIFY_SECRET | assert_empty

# Try removing the secret while it is still being used (should fail)
# Use or on lines that where the first half is expected to fail so script doesn't exit
docker secret rm $SECRET_NAME 2>err || echo "Removing secret while used fails as expected"
cat err | assert_contains "Error response from daemon"


# Remove access to the secret by running an update
docker service update  --detach=false --secret-rm $SECRET_NAME $SERVICE_NAME

# Check that update is completed
NUM_REPLICAS=$(check_replicas $SERVICE_NAME $REPLICAS $RETRIES)
assert_equals "Expect the service is up" $NUM_REPLICAS $REPLICAS

# Verify that there is no longer access to the secret
CONTAINER_ID=$(docker ps --filter name=$SERVICE_NAME -q)
docker exec $CONTAINER_ID cat /run/secrets/$SECRET_NAME 2>err || echo "Reading the secret the service fails as expected"
echo "The secret should be officially removed"
cat err | assert_contains "No such file or directory"

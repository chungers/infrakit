#/bin/sh
# Summary: Test Service deployment with stack deployment
# Labels:
# Repeat:


set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"

SERVICE_NAME="nginx_nginx"
STACK_NAME="nginx"
REPLICAS=3
RETRIES=5

clean_up(){
        docker stack rm $STACK_NAME
}

trap clean_up EXIT

# Deploy a stack
docker stack deploy --compose-file docker-stack.yml $STACK_NAME 

# Wait for the service to come up, and get the number of replicas
# Check replicas does a docker service ps and counts the output 
# since the service is created with a stack the name of the service in the
# compose file is combined with the name of stack
NUM_REPLICAS=$(check_replicas $SERVICE_NAME $REPLICAS $RETRIES) 

# Check the number of replicas
assert_equals "Expect the service to be replicated" $NUM_REPLICAS $REPLICAS 

# Check that the stack comes up under docker stack based cli 
docker stack ls | assert_contains "$STACK_NAME"
docker stack services $STACK_NAME | assert_contains "$STACK_NAME" 

exit 0 

#!/bin/sh
# Summary: Creates a service without an exposed port. Exposes a port and verifies that the service has that port exposed
# Lables:


REPS=3
NAME="ping3"

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"

# NAME replicas retries
check_replicas() {
REPLICAS=

for (( i = 1; i <= $3; i++ ))
do
   REPLICAS=$(docker service ps $1 | awk '{print $6}' | grep Running | wc -l)
   #echo $REPLICAS
   #echo Second parameter is $2    

    if [[ $REPLICAS -eq $2 ]];
        then break
    else
       sleep 5s
   fi
done

echo $REPLICAS

#if [[ $REPLICAS != $2 ]]; then
#    echo exiting with an error 
#    exit 1
#fi     
}


# Deploy a service with no ports exposed
docker service create --replicas $REPS --name $NAME alpine ping docker.com

# Check that the service is up and running (make this a library function later in some way or other)
ACTUAL=$(check_replicas $NAME $REPS 10)
echo "Acutal number of replicas $ACTUAL expected number of replicas $REPS"
assert_equals "Correct number of replicas" $ACTUAL $REPS 

# Check that service has no ports exposed
ACTUAL=$(docker service ls | grep $NAME | awk '{ print $6 }' | wc -w)
assert_equals "service has no ports exposed" $ACTUAL 0  

# Expose a port
EXPOSED_PORT=8000
docker service update --publish-add $EXPOSED_PORT $NAME 

# Check that port is available
ACUTAL=$(docker service ls | grep $NAME | awk '{ print $6 }')
assert_contains "Service has  ports exposed" $ACTUAL $EXPOSED_PORT

# Delete the port
docker service update --publish-rm $EXPOSED_PORT $NAME 

# Check that the port is not available
ACTUAL=$(docker service ls | grep $NAME | awk '{ print $6 }' | wc -w)
assert_equals "service has no ports exposed" $ACTUAL 0  

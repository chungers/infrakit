#!/bin/sh
# SUMMARY: Check that all the helper containers are up and running for docker for azure or docker for aws
# LABELS: editions 

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"


EDITIONS_CONTAINERS="controller meta guide"
AZURE_CONTAINERS="logger agent"
AWS_CONTAINERS="shell"
PLATFORM=

if [ $VERSION -z ]; then
    VERSION="17.06"
fi


# Function that checks both for a container name
# and if it is up
check_container() {
  CONTAINER=$(docker container ls --filter name=$1 --filter status="running")
  echo $CONTAINER | assert_contains $1
  echo $CONTAINER | assert_contains $VERSION
}

# Determine whether the platform is Azure or AWS
AWS=$(docker container ls | grep "aws") || AZURE=$(docker container ls | grep "azure") 

if [ -z "$AWS" ]; 
        then
        PLATFORM="azure"  
        EDITIONS_CONTAINERS=$EDITIONS_CONTAINERS" "$AZURE_CONTAINERS
elif [ -z "$AZURE" ];
        then
        PLATFORM="aws" 
        EDITIONS_CONTAINERS=$EDITIONS_CONTAINERS" "$AWS_CONTAINERS
fi


#Check that the necessary containers are up and running
for container in $EDITIONS_CONTAINERS 
do
        check_container $container
done

#Check that cloudstor exists and that enabled is true 
docker plugin ls | assert_contains "cloudstor:$PLATFORM"
docker plugin ls --filter "enabled=true"  --filter 'capability=volumedriver' | assert_contains "cloudstor:$PLATFORM"

#!/bin/bash
echo "#================"
echo "Start DDC setup"

# default to yes if INSTALL DDC is empty.
INSTALL_DDC=${INSTALL_DDC:-"yes"}

PRODUCTION_HUB_NAMESPACE='docker'
HUB_NAMESPACE=${HUB_NAMESPACE:-"dockerorcadev"}
HUB_TAG=${HUB_TAG-"2.0.0-beta1"}
IMAGE_LIST_ARGS=''

echo "PATH=$PATH"
echo "STACK_NAME=$STACK_NAME"
echo "AWS_REGION=$REGION"
echo "INSTALL_DDC=$INSTALL_DDC"
echo "ELB_NAME=$ELB_NAME"
echo "UCP_ADMIN_USER=$UCP_ADMIN_USER"
echo "UCP_IMAGE=${HUB_NAMESPACE}/ucp:${HUB_TAG}"
echo "#================"

# we don't want to install, exit now.
if [[ "$INSTALL_DDC" != "yes" ]] ; then
    exit 0
fi

if [[ "$HUB_NAMESPACE" != "$PRODUCTION_HUB_NAMESPACE" ]]; then
    IMAGE_LIST_ARGS=" --image-version dev: "
fi

echo "Load the docker images"
if [[ -n "$DOCKER_ID" && -n "$DOCKER_ID_PASSWORD" ]]; then
    echo "Logging in to Hub account $DOCKER_ID"
    # TODO should we check for login success here?
    docker login -u "$DOCKER_ID" -p "$DOCKER_ID_PASSWORD"
fi

# TODO this only works on images built since Orca commit
# https://github.com/docker/orca/commit/78323da280160f67eae6329f3cb4bdd7c6a83bf9
# So e.g., dockerorcadev/ucp:2.0.0-tp1 won't install with this mechanism as of 31 August.
# When this code is used, the for loop below can be removed.
#docker run --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock \
#       "${HUB_NAMESPACE}/ucp:${HUB_TAG}" images \
#       -D \
#       --pull always \
#       --registry-username $DOCKER_ID \
#       --registry-password "$DOCKER_ID_PASSWORD" \
#       $IMAGE_LIST_ARGS
images=$(docker run --rm "${HUB_NAMESPACE}/ucp:${HUB_TAG}" images --list $IMAGE_LIST_ARGS )
for im in $images; do
    docker pull $im
done

if [ "$NODE_TYPE" == "worker" ] ; then
     # nothing left to do for workers, so exit.
     exit 0
fi

echo "Wait until CloudFormation is complete"
time aws cloudformation wait stack-create-complete --stack-name ${STACK_NAME} --region=$REGION
echo "CloudFormation is complete, time to proceed."

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
    echo "We are the swarm leader"
    echo "Setup DDC"

    SSH_ELB_PHYS_ID=$(aws cloudformation describe-stack-resources --stack-name ${STACK_NAME} --region=$REGION --logical-resource-id $ELB_NAME | jq -r ".StackResources[0].PhysicalResourceId")
    echo "SSH_ELB_PHYS_ID=$SSH_ELB_PHYS_ID"
    # Add port 443 since we'll need it later...
    aws elb create-load-balancer-listeners --region $REGION --load-balancer-name ${SSH_ELB_PHYS_ID} --listeners "Protocol=TCP,LoadBalancerPort=443,InstanceProtocol=TCP,InstancePort=443"

    # Get All Load Balancers
    # read lb1 lb2 <<< $(aws cloudformation describe-stack-resources --stack-name ${STACK_NAME} --region=$REGION | jq '.StackResources[] | select(.LogicalResourceId | endswith("LoadBalancer")) | .PhysicalResourceId')
    # Get All Load Balancers DNS Name
    read lb1 lb2 <<< $(aws elb describe-load-balancers --region=$REGION | jq ".LoadBalancerDescriptions[] | select(.DNSName | startswith(\"${STACK_NAME}-ELB\")) | .DNSName")
    if [ $lb1 == *"-ELB-SSH" ]
    then
        ELB_HOSTNAME=$lb2
        SSH_ELB_HOSTNAME=$lb1
    else
        ELB_HOSTNAME=$lb1
        SSH_ELB_HOSTNAME=$lb2
    fi

    echo "ELB_HOSTNAME=$ELB_HOSTNAME"
    echo "SSH_ELB_HOSTNAME=$SSH_ELB_HOSTNAME"

    echo "Run the DDC install script"
    docker run --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock ${HUB_NAMESPACE}/ucp:${HUB_TAG} install --san "$SSH_ELB_HOSTNAME" --external-service-lb "$ELB_HOSTNAME" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
    echo "Finished"
else
    echo "Not the swarm leader, nothing to do, exiting"
fi

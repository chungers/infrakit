#!/bin/bash
echo "#================"
echo "Start DDC setup"

PATH=$PATH:/usr/docker/bin

# default to yes if INSTALL DDC is empty.
INSTALL_DDC=${INSTALL_DDC:-"yes"}

echo "PATH=$PATH"
echo "STACK_NAME=$STACK_NAME"
echo "AWS_REGION=$AWS_DEFAULT_REGION"
echo "INSTALL_DDC=$INSTALL_DDC"
echo "ELB_NAME=$ELB_NAME"
echo "#================"

# we don't want to install, exit now.
if [[ "$INSTALL_DDC" != "yes" ]] ; then
    exit 0
fi

echo "Load the docker images"
wget -qO- https://s3.amazonaws.com/packages.docker.com/caas/79Az36QAF4WGuvZdcJ7T/ucp_images_1.2.0-alpha1.tar.gz | docker load

if [ "$NODE_TYPE" == "worker" ] ; then
     # nothing left to do for workers, so exit.
     exit 0
fi

echo "Wait until CloudFormation is complete"
time aws cloudformation wait stack-create-complete --stack-name ${STACK_NAME} --region=$AWS_DEFAULT_REGION
echo "CloudFormation is complete, time to proceed."

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
    echo "We are the swarm leader"
    echo "Setup DDC"

    SSH_ELB_PHYS_ID=$(aws cloudformation describe-stack-resources --stack-name ${STACK_NAME} --region=$AWS_DEFAULT_REGION --logical-resource-id $ELB_NAME | jq -r ".StackResources[0].PhysicalResourceId")

    echo "SSH_ELB_PHYS_ID=$SSH_ELB_PHYS_ID"
    SSH_ELB_HOSTNAME=$(aws elb describe-load-balancers --load-balancer-names ${SSH_ELB_PHYS_ID} --region=$AWS_DEFAULT_REGION | jq -r ".LoadBalancerDescriptions[0].DNSName")
    echo "SSH_ELB_HOSTNAME=$SSH_ELB_HOSTNAME"
    # Add port 443 since we'll need it later...
    aws elb create-load-balancer-listeners --load-balancer-name ${SSH_ELB_PHYS_ID} --listeners "Protocol=TCP,LoadBalancerPort=443,InstanceProtocol=TCP,InstancePort=443"

    echo "Run the DDC install script"
    docker run --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock dockerorcadev/ucp:1.2.0-alpha1 install --image-version dev: --san $SSH_ELB_HOSTNAME
    echo "Finished"
else
    echo "Not the swarm leader, nothing to do, exiting"
fi

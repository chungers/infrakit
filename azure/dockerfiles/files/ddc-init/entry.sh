#!/bin/bash
echo "#================"
echo "Start DDC setup"

echo "PATH=$PATH"
echo "ROLE=$ROLE"
echo "DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION"
echo "ACCOUNT_ID=$ACCOUNT_ID"
echo "REGION=$REGION"
echo "ELB_NAME=$ELB_NAME"
echo "UCP_ADMIN_USER=$UCP_ADMIN_USER"
echo "APP_ID=$APP_ID"
echo "TENANT_ID=$TENANT_ID"
echo "#================"

echo "Load the docker images"
wget -qO- https://s3.amazonaws.com/packages.docker.com/caas/79Az36QAF4WGuvZdcJ7T/ucp_images_2.0.0-tp1 | docker load

if [ "$NODE_TYPE" == "worker" ] ; then
     # nothing left to do for workers, so exit.
     exit 0
fi

echo "Wait until Resource Group is complete"
# Login via the service principal
azure login -u "${APP_ID}" -p "${APP_PASS}" --service-principal --tenant "${TENANT_ID}"
 
while :
do
	provisioning_state=$(azure group deployment list ${STACK_NAME} --json | jq -r '.[0] | .properties.provisioningState')
	if [ $provisioning_state = "Success" ]
	then
		break
	elif [ $provisioning_state = "Failed" ]
	then
		echo "Resource group provisioning failed"
		exit 0
	fi
	echo "."
done
time 
echo "Resource Group is complete, time to proceed."
# 
IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
    echo "We are the swarm leader"
    echo "Setup DDC"

    # SSH_ELB_PHYS_ID=$(azure group show ${STACK_NAME} --json | jq -r '.resources | .[] | select(.name=="externalLoadBalancer") | .id')
    SSH_ELB_NAME=$(azure group show ${STACK_NAME} --json | jq -r '.resources | .[] | select(.name=="${SSH_ELB_NAME}") | .name')

    # echo "SSH_ELB_PHYS_ID=$SSH_ELB_PHYS_ID"
    # SSH_ELB_HOSTNAME=$(aws elb describe-load-balancers --load-balancer-names ${SSH_ELB_PHYS_ID} --region=$REGION | jq -r ".LoadBalancerDescriptions[0].DNSName")
    # echo "SSH_ELB_HOSTNAME=$SSH_ELB_HOSTNAME"
    
    # aws elb create-load-balancer-listeners --region $REGION --load-balancer-name ${SSH_ELB_PHYS_ID} --listeners "Protocol=TCP,LoadBalancerPort=443,InstanceProtocol=TCP,InstancePort=443"
    SSH_ELB_HOSTNAME=$(azure resource show ${STACK_NAME} ${SSH_ELB_NAME} "Microsoft.Network/loadBalancers" "2016-09-01")

    # # Add port 443 since we'll need it later...
    #azure network lb inbound-nat-rule create -g ${STACK_NAME} -l ${SSH_ELB_NAME} -n ddc1 -p tcp -f 443 -b 443
    # azure network lb rule create ${STACK_NAME} ${SSH_ELB_NAME} lbrule -p tcp -f 443 -b 443 -t DDCfrontendpool -o DDCbackendpool


    echo "Run the DDC install script"
    docker run --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock docker/ucp:2.0.0-tp1 install --san $SSH_ELB_HOSTNAME --admin-username $UCP_ADMIN_USER --admin-password $UCP_ADMIN_PASSWORD
    echo "Finished"
else
    echo "Not the swarm leader, nothing to do, exiting"
fi

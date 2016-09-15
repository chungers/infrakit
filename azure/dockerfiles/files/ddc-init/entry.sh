#!/bin/bash
echo "#================"
echo "Start DDC setup"

echo "PATH=$PATH"
echo "ROLE=$ROLE"
echo "REGION=$REGION"
echo "RGROUP_NAME=$RGROUP_NAME"
echo "LB_NAME=$LB_NAME"
echo "LB_IP=$LB_IP"
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
azure login -u "${APP_ID}" -p "${APP_SECRET}" --service-principal --tenant "${TENANT_ID}"
if [[ $? -ne 0 ]]
then
	exit 0
fi

# FIX CAPTURE OF BAD LOGIN !!!!


COUNTER=0
while :
do
	provisioning_state=$(azure group deployment list ${RGROUP_NAME} --json | jq -r '.[0] | .properties.provisioningState')
	if [ "$provisioning_state" == "Succeeded" ]
	then
		break
	elif [ "$provisioning_state" == "Failed" ]
	then
		echo "Resource group provisioning failed"
		exit 0
	fi
	echo "."
	COUNTER=$((COUNTER + 1))
	if [ $COUNTER -gt 10000 ]
	then
		echo "Resource group setup status unknown"
		exit 0
	fi
done
time 
echo "Resource Group is complete, time to proceed."
# 
IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
	echo "We are the swarm leader"
	echo "Setup DDC"

	# SSH_LB_PHYS_IDAME=$(azure group show ${RGROUP_NAME} --json | jq -r '.resources | .[] | select(.name=="${SSH_ELB_NAME}") | .name')
	# SSH_LB_ID=$(azure resource show ${RGROUP_NAME} ${SSH_ELB_NAME} "Microsoft.Network/loadBalancers" "2016-09-01" --json | jq -r '.properties.frontendIPConfigurations[0].properties.publicIPAddress.id')
	# SSH_LB_NAME=${SSH_LB_ID##*/}
	SSH_LB_IP=$(azure network public-ip show ${RGROUP_NAME} ${LB_NAME} --json | jq -r '.ipAddress')

	echo "Run the DDC install script"
	docker run --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock docker/ucp:2.0.0-tp1 install --san $LB_IP --admin-username $UCP_ADMIN_USER --admin-password $UCP_ADMIN_PASSWORD
	echo "Finished"
else
	echo "Not the swarm leader, nothing to do, exiting"
fi

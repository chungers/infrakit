#!/bin/bash
echo "#================"
echo "Start DDC setup"


PRODUCTION_HUB_NAMESPACE='docker'
HUB_NAMESPACE=${HUB_NAMESPACE:-"docker"}
HUB_TAG=${HUB_TAG-"2.0.0-beta3"}
IMAGE_LIST_ARGS=''

echo "PATH=$PATH"
echo "ROLE=$ROLE"
echo "REGION=$REGION"
echo "RGROUP_NAME=$RGROUP_NAME"
echo "APP_ID=$APP_ID"
echo "TENANT_ID=$TENANT_ID"
echo "LB_NAME=$LB_NAME"
echo "LB_IP=$LB_IP"
echo "UCP_ADMIN_USER=$UCP_ADMIN_USER"
echo "UCP_IMAGE=${HUB_NAMESPACE}/ucp:${HUB_TAG}"
echo "#================"


if [[ "$HUB_NAMESPACE" != "$PRODUCTION_HUB_NAMESPACE" ]]; then
    IMAGE_LIST_ARGS=" --image-version dev: "
fi

echo "Load the docker images"
images=$(docker run --rm "${HUB_NAMESPACE}/ucp:${HUB_TAG}" images --list $IMAGE_LIST_ARGS )
for im in $images; do
    docker pull $im
done

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

 
IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
	echo "We are the swarm leader"
	echo "Setup DDC"

	read lb1 lb2 <<< $(azure group show ${RGROUP_NAME} --json | jq -r '.resources | .[] | select(.id | endswith("LoadBalancer-public-ip")) | .id')
	if [ $lb1 == *"SSHLoadBalancer-public-ip" ]
	then
			LB_IP=$(azure network public-ip show ${RGROUP_NAME} ${lb2##*/} --json | jq -r '.ipAddress')
			SSH_LB_IP=$(azure network public-ip show ${RGROUP_NAME} ${lb1##*/} --json | jq -r '.ipAddress')
	else
			LB_IP=$(azure network public-ip show ${RGROUP_NAME} ${lb1##*/} --json | jq -r '.ipAddress')
			SSH_LB_IP=$(azure network public-ip show ${RGROUP_NAME} ${lb2##*/} --json | jq -r '.ipAddress')
	fi
	echo "LB_IP=$LB_IP"
	echo "SSH_LB_IP=$SSH_LB_IP"

	echo "Run the DDC install script"
	docker run --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock ${HUB_NAMESPACE}/ucp:${HUB_TAG} install --san "$SSH_LB_IP" --external-service-lb "$LB_IP" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
	echo "Finished"
else
	echo "Not the swarm leader, nothing to do, exiting"
fi

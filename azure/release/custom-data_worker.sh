export ROLE="WORKER"
export REGION="variables('storageLocation')"
export ACCOUNT_ID="variables('accountID')"
export DOCKER_FOR_IAAS_VERSION="variables('dockerForIAASVersion')"

docker run --restart=no -d -e ROLE=$ROLE -e REGION=$REGION -e ACCOUNT_ID=$ACCOUNT_ID -e MANAGER_IP=$MANAGER_IP -e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/init-azure:$DOCKER_FOR_IAAS_VERSION
docker run --restart=always -d -e ROLE=$ROLE -e REGION=$REGION -e ACCOUNT_ID=$ACCOUNT_ID -e MANAGER_IP=$MANAGER_IP -e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/guide-azure:$DOCKER_FOR_IAAS_VERSION

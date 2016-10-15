export ROLE="MANAGER"
export ACCOUNT_ID="variables('accountID')"
export DOCKER_FOR_IAAS_VERSION="variables('dockerForIAASVersion')"
export REGION="variables('storageLocation')"
export SUB_ID="variables('accountID')"
export GROUP_NAME="variables('groupName')"
export LB_NAME="variables('lbName')"
export APP_ID="variables('adServicePrincipalAppID')"
export APP_SECRET="variables('adServicePrincipalAppSecret')"
export CLOUD_USER="variables('dockerCloudUsername')"
export CLOUD_KEY="variables('dockerCloudAPIKey')"
export SWARM_NAME="variables('dockerCloudClusterName')"
 

docker run --restart=no -d -e ROLE="$ROLE" -e REGION="$REGION" -e ACCOUNT_ID="$ACCOUNT_ID" -e PRIVATE_IP="$MANAGER_IP" -e DOCKER_FOR_IAAS_VERSION="$DOCKER_FOR_IAAS_VERSION" -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/init-azure:"$DOCKER_FOR_IAAS_VERSION"
docker run --restart=always -d -e ROLE="$ROLE -e REGION="$REGION -e ACCOUNT_ID="$ACCOUNT_ID" -e PRIVATE_IP="$MANAGER_IP" -e DOCKER_FOR_IAAS_VERSION="$DOCKER_FOR_IAAS_VERSION" -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/guide-azure:"$DOCKER_FOR_IAAS_VERSION"
echo default: "$LB_NAME" >> /var/lib/docker/swarm/elb.config
echo "$LB_NAME" > /var/lib/docker/swarm/lb_name
docker run -v /var/run/docker.sock:/var/run/docker.sock  -v /var/lib/docker/swarm:/var/lib/docker/swarm --name=editions_controller docker4x/l4controller-azure:"$DOCKER_FOR_IAAS_VERSION" run --ad_app_id="$APP_ID" --ad_app_secret="$APP_SECRET" --subscription_id="$SUB_ID" --resource_group="$GROUP_NAME" --log=4 --default_lb_name="$LB_NAME" --environment=AzurePublicCloud



export IS_LEADER=$(docker node inspect self -f "{{ .ManagerStatus.Leader }}") 
if [ "$IS_LEADER" == "true" ]; then 
docker run --rm --name=cloud_registration -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_USER="$CLOUD_USER" -e DOCKER_PASS="$CLOUD_KEY" -e SWARM_NAME="$SWARM_NAME" -e INTERNAL_ENDPOINT="$LB_SSH_IP:2376" docker4x/cloud-azure:$DOCKER_FOR_IAAS_VERSION
fi

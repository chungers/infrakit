export ROLE="MANAGER"
export ACCOUNT_ID="variables('accountID')"
export DOCKER_FOR_IAAS_VERSION="variables('dockerForIAASVersion')"
export REGION="variables('storageLocation')"
export SUB_ID="variables('accountID')"
export GROUP_NAME="variables('groupName')"
export LB_NAME="variables('lbName')"
export APP_ID="variables('adServicePrincipalAppID')"
export APP_SECRET="variables('adServicePrincipalAppSecret')"
export TENANT_ID="variables('adServicePrincipalTenantID')"

export DDC_USER="variables('ddcUser')"
export DDC_PASS="variables('ddcPass')"
export RGROUP_NAME="variables('groupName')"; 
export LB_NAME="variables('lbPublicIPAddressName')"; 
export LB_IP="variables('lbPublicIPAddress')";

 

docker run --restart=no -d -e ROLE="$ROLE" -e REGION="$REGION" -e ACCOUNT_ID="$ACCOUNT_ID" -e PRIVATE_IP="$MANAGER_IP" -e DOCKER_FOR_IAAS_VERSION="$DOCKER_FOR_IAAS_VERSION" -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/init-azure:"$DOCKER_FOR_IAAS_VERSION"
docker run --restart=always -d -e ROLE="$ROLE -e REGION="$REGION -e ACCOUNT_ID="$ACCOUNT_ID" -e PRIVATE_IP="$MANAGER_IP" -e DOCKER_FOR_IAAS_VERSION="$DOCKER_FOR_IAAS_VERSION" -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/guide-azure:"$DOCKER_FOR_IAAS_VERSION"
echo default: "$LB_NAME" >> /var/lib/docker/swarm/elb.config
echo "$LB_NAME" > /var/lib/docker/swarm/lb_name
docker run -v /var/run/docker.sock:/var/run/docker.sock  -v /var/lib/docker/swarm:/var/lib/docker/swarm --name=editions_controller docker4x/l4controller-azure:"$DOCKER_FOR_IAAS_VERSION" run --ad_app_id="$APP_ID" --ad_app_secret="$APP_SECRET" --subscription_id="$SUB_ID" --resource_group="$GROUP_NAME" --log=4 --default_lb_name="$LB_NAME" --environment=AzurePublicCloud


docker run --restart=no --rm -e ROLE=$ROLE -e REGION=$REGION -e ACCOUNT_ID=$ACCOUNT_ID -e APP_ID=$APP_ID -e APP_SECRET=$APP_SECRET -e TENANT_ID=$TENANT_ID -e RGROUP_NAME=$RGROUP_NAME -e UCP_ADMIN_USER=$DDC_USER -e UCP_ADMIN_PASSWORD=$DDC_PASS -e LB_NAME=$LB_NAME -e LB_IP=$LB_SSH_IP -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/ddc-init-azure:$DOCKER_FOR_IAAS_VERSION


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
export SWARM_INFO_TABLE="variables('swarmInfoTable')"
export SWARM_INFO_STORAGE_ACCOUNT="variables('swarmInfoStorageAccount')"
export SWARM_LOGS_STORAGE_ACCOUNT="variables('swarmLogsStorageAccount')"


# create daemon config with custom tag
echo "{\"log-driver\": \"syslog\",\"log-opts\": {\"syslog-address\": \"udp://localhost:514\", \"tag\": \"{{.Name}}/{{.ID}}\" }}" > /etc/docker/daemon.json
service docker restart
sleep 5

# add logging container
docker volume create --name container-logs
docker run --log-driver=json-file --name=editions_logger --restart=always -d -e ROLE="$ROLE" -e REGION="$REGION" -e TENANT_ID="$TENANT_ID" -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e ACCOUNT_ID="$ACCOUNT_ID" -e GROUP_NAME="$GROUP_NAME" -e SWARM_LOGS_STORAGE_ACCOUNT="$SWARM_LOGS_STORAGE_ACCOUNT" -e SWARM_FILE_SHARE=`hostname` -p 514:514/udp -v container-logs:/log/ docker4x/logger-azure:"$DOCKER_FOR_IAAS_VERSION"

docker run --log-driver=json-file  --restart=no -d -e ROLE="$ROLE" -e REGION="$REGION" -e TENANT_ID="$TENANT_ID" -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e ACCOUNT_ID="$ACCOUNT_ID" -e GROUP_NAME="$GROUP_NAME" -e PRIVATE_IP="$MANAGER_IP" -e DOCKER_FOR_IAAS_VERSION="$DOCKER_FOR_IAAS_VERSION" -e SWARM_INFO_TABLE="$SWARM_INFO_TABLE" -e SWARM_INFO_STORAGE_ACCOUNT="$SWARM_INFO_STORAGE_ACCOUNT" -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/init-azure:"$DOCKER_FOR_IAAS_VERSION"
docker run --log-driver=json-file  --restart=always -d -e ROLE="$ROLE" -e REGION="$REGION" -e TENANT_ID="$TENANT_ID" -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e ACCOUNT_ID="$ACCOUNT_ID" -e GROUP_NAME="$GROUP_NAME" -e PRIVATE_IP="$MANAGER_IP" -e DOCKER_FOR_IAAS_VERSION="$DOCKER_FOR_IAAS_VERSION" -e SWARM_INFO_TABLE="$SWARM_INFO_TABLE" -e SWARM_INFO_STORAGE_ACCOUNT="$SWARM_INFO_STORAGE_ACCOUNT" -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/guide-azure:"$DOCKER_FOR_IAAS_VERSION"
echo default: "$LB_NAME" >> /var/lib/docker/swarm/elb.config
echo "$LB_NAME" > /var/lib/docker/swarm/lb_name
docker run --log-driver=json-file  -v /var/run/docker.sock:/var/run/docker.sock  -v /var/lib/docker/swarm:/var/lib/docker/swarm --name=editions_controller docker4x/l4controller-azure:"$DOCKER_FOR_IAAS_VERSION" run --ad_app_id="$APP_ID" --ad_app_secret="$APP_SECRET" --subscription_id="$SUB_ID" --resource_group="$GROUP_NAME" --log=4 --default_lb_name="$LB_NAME" --environment=AzurePublicCloud

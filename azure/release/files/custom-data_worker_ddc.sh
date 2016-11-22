export ROLE="WORKER"
export REGION="variables('storageLocation')"
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
export MANAGER_IP=$(ifconfig eth0 | grep "inet addr:" | cut -d: -f2 | cut -d" " -f1)

# create daemon config with custom tag
echo "{\"log-driver\": \"syslog\",\"log-opts\": {\"syslog-address\": \"udp://localhost:514\", \"tag\": \"{{.Name}}/{{.ID}}\" }}" > /etc/docker/daemon.json
service docker restart
sleep 5

# add logging container
docker volume create --name container-logs
docker run --log-driver=json-file --name=editions_logger --restart=always -d -e ROLE="$ROLE" -e REGION="$REGION" -e TENANT_ID="$TENANT_ID" -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e ACCOUNT_ID="$ACCOUNT_ID" -e GROUP_NAME="$GROUP_NAME" -e SWARM_LOGS_STORAGE_ACCOUNT="$SWARM_LOGS_STORAGE_ACCOUNT" -e SWARM_FILE_SHARE=`hostname` -p 514:514/udp -v container-logs:/log/ docker4x/logger-azure:"$DOCKER_FOR_IAAS_VERSION"
# init container
docker run --log-driver=json-file --restart=no -d -e ROLE=$ROLE -e REGION=$REGION -e TENANT_ID="$TENANT_ID" -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e GROUP_NAME="$GROUP_NAME" -e SWARM_INFO_TABLE="$SWARM_INFO_TABLE" -e SWARM_INFO_STORAGE_ACCOUNT="$SWARM_INFO_STORAGE_ACCOUNT" -e ACCOUNT_ID=$ACCOUNT_ID -e MANAGER_IP=$MANAGER_IP -e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/init-azure:$DOCKER_FOR_IAAS_VERSION
# guide container
docker run --log-driver=json-file --restart=always -d -e ROLE=$ROLE -e REGION=$REGION -e ACCOUNT_ID=$ACCOUNT_ID -e MANAGER_IP=$MANAGER_IP -e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/guide-azure:$DOCKER_FOR_IAAS_VERSION
# ddc-init container
docker run --log-driver=json-file --restart=no --rm -e ROLE=$ROLE -e REGION=$REGION -e ACCOUNT_ID=$ACCOUNT_ID -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/ddc-init-azure:$DOCKER_FOR_IAAS_VERSION

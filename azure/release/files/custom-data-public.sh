export ACCOUNT_ID="variables('accountID')"
export REGION="variables('storageLocation')"
export SUB_ID="variables('accountID')"
export GROUP_NAME="variables('groupName')"
export LB_NAME="variables('lbName')"
export APP_ID="variables('adServicePrincipalAppID')"
export APP_SECRET="variables('adServicePrincipalAppSecret')"
export TENANT_ID="variables('adServicePrincipalTenantID')"
export SWARM_INFO_STORAGE_ACCOUNT="variables('swarmInfoStorageAccount')"
export SWARM_LOGS_STORAGE_ACCOUNT="variables('swarmLogsStorageAccount')"
export PRIVATE_IP=$(ifconfig eth0 | grep "inet addr:" | cut -d: -f2 | cut -d" " -f1)
export AZURE_HOSTNAME=$(hostname)

docker run --label com.docker.editions.system --log-driver=json-file --restart=no -it -e LB_NAME -e SUB_ID -e ROLE -e REGION -e TENANT_ID -e APP_ID -e APP_SECRET -e ACCOUNT_ID -e GROUP_NAME -e PRIVATE_IP -e DOCKER_FOR_IAAS_VERSION -e SWARM_INFO_STORAGE_ACCOUNT -e SWARM_LOGS_STORAGE_ACCOUNT -e AZURE_HOSTNAME -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/lib/docker:/var/lib/docker -v /var/log:/var/log docker4x/init-azure:"$DOCKER_FOR_IAAS_VERSION"

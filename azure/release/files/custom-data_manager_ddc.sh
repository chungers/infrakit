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
export MANAGER_IP=$(ifconfig eth0 | grep "inet addr:" | cut -d: -f2 | cut -d" " -f1)
export VMSS_MGR="dockerswarm-managervmss"
export VMSS_WRK="dockerswarm-worker-vmss"
# DDC Specific
export DDC_USER="variables('ddcUser')"
export DDC_PASS="variables('ddcPass')"

# create daemon config with custom tag
echo "{\"log-driver\": \"syslog\",\"log-opts\": {\"syslog-address\": \"udp://localhost:514\", \"tag\": \"{{.Name}}/{{.ID}}\" }}" > /etc/docker/daemon.json
service docker restart
sleep 5

# add logging container
docker volume create --name container-logs
docker run --log-driver=json-file --name=editions_logger --restart=always -d -e ROLE="$ROLE" -e REGION="$REGION" -e TENANT_ID="$TENANT_ID" -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e ACCOUNT_ID="$ACCOUNT_ID" -e GROUP_NAME="$GROUP_NAME" -e SWARM_LOGS_STORAGE_ACCOUNT="$SWARM_LOGS_STORAGE_ACCOUNT" -e SWARM_FILE_SHARE=`hostname` -p 514:514/udp -v container-logs:/log/ docker4x/logger-azure:"$DOCKER_FOR_IAAS_VERSION"
# token server
docker run --log-driver=json-file --name=meta-azure --restart=always -d -p $MANAGER_IP:9024:8080 -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e SUBSCRIPTION_ID="$SUB_ID" -e TENANT_ID="$TENANT_ID" -e GROUP_NAME="$GROUP_NAME" -e VMSS_MGR="$VMSS_MGR" -e VMSS_WRK="$VMSS_WRK" -v /var/run/docker.sock:/var/run/docker.sock docker4x/meta-aws:$DOCKER_FOR_IAAS_VERSION metaserver -flavor=azure
# init container
docker run --log-driver=json-file  --restart=no -d -e ROLE="$ROLE" -e REGION="$REGION" -e TENANT_ID="$TENANT_ID" -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e ACCOUNT_ID="$ACCOUNT_ID" -e GROUP_NAME="$GROUP_NAME" -e PRIVATE_IP="$MANAGER_IP" -e DOCKER_FOR_IAAS_VERSION="$DOCKER_FOR_IAAS_VERSION" -e SWARM_INFO_TABLE="$SWARM_INFO_TABLE" -e SWARM_INFO_STORAGE_ACCOUNT="$SWARM_INFO_STORAGE_ACCOUNT" -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/init-azure:"$DOCKER_FOR_IAAS_VERSION"
# guid container
docker run --log-driver=json-file  --restart=always -d -e ROLE="$ROLE" -e REGION="$REGION" -e TENANT_ID="$TENANT_ID" -e APP_ID="$APP_ID" -e APP_SECRET="$APP_SECRET" -e ACCOUNT_ID="$ACCOUNT_ID" -e GROUP_NAME="$GROUP_NAME" -e PRIVATE_IP="$MANAGER_IP" -e DOCKER_FOR_IAAS_VERSION="$DOCKER_FOR_IAAS_VERSION" -e SWARM_INFO_TABLE="$SWARM_INFO_TABLE" -e SWARM_INFO_STORAGE_ACCOUNT="$SWARM_INFO_STORAGE_ACCOUNT" -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/guide-azure:"$DOCKER_FOR_IAAS_VERSION"
echo default: "$LB_NAME" >> /var/lib/docker/swarm/elb.config
echo "$LB_NAME" > /var/lib/docker/swarm/lb_name
# ddc init container
docker run --log-driver=json-file --restart=no --rm -e ROLE=$ROLE -e REGION=$REGION -e ACCOUNT_ID=$ACCOUNT_ID -e APP_ID=$APP_ID -e APP_SECRET=$APP_SECRET -e TENANT_ID=$TENANT_ID -e RGROUP_NAME=$RGROUP_NAME -e UCP_ADMIN_USER=$DDC_USER -e UCP_ADMIN_PASSWORD=$DDC_PASS -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/ddc-init-azure:$DOCKER_FOR_IAAS_VERSION
# l4controller container
docker run -d --log-driver=json-file -v /var/run/docker.sock:/var/run/docker.sock -v /var/lib/docker/swarm:/var/lib/docker/swarm --name=editions_controller docker4x/l4controller-azure:"$DOCKER_FOR_IAAS_VERSION" run --ad_app_id="$APP_ID" --ad_app_secret="$APP_SECRET" --subscription_id="$SUB_ID" --resource_group="$GROUP_NAME" --log=4 --default_lb_name="$LB_NAME" --environment=AzurePublicCloud


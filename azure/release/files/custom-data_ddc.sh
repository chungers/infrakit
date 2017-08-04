export MANAGER_COUNT="variables('managerCount')"
export WORKER_COUNT="variables('workerCount')"
export DTR_STORAGE_ACCOUNT="variables('dtrStorageAccount')"
export UCP_ADMIN_USER="variables('ddcUser')"
export UCP_ADMIN_PASSWORD="variables('ddcPass')"
export UCP_LICENSE="variables('ddcLicense')"
export APP_ELB_HOSTNAME="variables('extlbname')"
export UCP_ELB_HOSTNAME="variables('ucplbname')"
export DTR_ELB_HOSTNAME="variables('dtrlbname')"
export UCP_TAG="variables('ucpTag')"
export DTR_TAG="variables('dtrTag')"

docker run --log-driver=json-file --restart=on-failure:5 -d -e LB_NAME -e SUB_ID -e ROLE -e REGION -e TENANT_ID -e APP_ID -e APP_SECRET -e ACCOUNT_ID -e GROUP_NAME -e PRIVATE_IP -e DOCKER_FOR_IAAS_VERSION -e SWARM_INFO_STORAGE_ACCOUNT -e SWARM_LOGS_STORAGE_ACCOUNT -e RESOURCE_MANAGER_ENDPOINT -e SERVICE_MANAGEMENT_ENDPOINT -e ACTIVE_DIRECTORY_ENDPOINT -e STORAGE_ENDPOINT -e AZURE_HOSTNAME -e APP_ELB_HOSTNAME -e UCP_ELB_HOSTNAME -e DTR_ELB_HOSTNAME -e DTR_STORAGE_ACCOUNT -e MANAGER_COUNT -e WORKER_COUNT -e UCP_TAG -e DTR_TAG -e UCP_ADMIN_USER -e UCP_ADMIN_PASSWORD -e UCP_LICENSE -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/lib/docker:/var/lib/docker -v /var/log:/var/log -v /tmp/docker:/tmp/docker docker4x/ddc-init-azure:$DOCKER_FOR_IAAS_VERSION

export SWARM_NAME="variables('dockerCloudClusterName')"
export DOCKER_USER="variables('dockerCloudUsername')"
export DOCKER_PASS="variables('dockerCloudAPIKey')"
export DOCKERCLOUD_REST_HOST="variables('dockerCloudRestHost')"
export JWT_URL="variables('dockerIDJWTURL')"
export JWK_URL="variables('dockerIDJWKURL')"

# cloud registration container
export IS_LEADER=$(docker node inspect self -f "{{ .ManagerStatus.Leader }}") 
if [ "$IS_LEADER" == "true" ]; then
docker run --rm --name=cloud_registration -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_USER -e DOCKER_PASS -e SWARM_NAME -e INTERNAL_ENDPOINT="$LB_SSH_IP" -e DOCKERCLOUD_REST_HOST -e JWT_URL -e JWK_URL docker4x/cloud-azure:$DOCKER_FOR_IAAS_VERSION
fi

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')
if [ "$IS_LEADER" == "true" ]; then
  docker run --label com.docker.editions.system --log-driver=json-file --name=cloud-aws --rm --name=cloud_registration -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_USER=$DOCKERCLOUD_USER -e DOCKER_PASS=$DOCKERCLOUD_API_KEY -e SWARM_NAME=$SWARM_NAME -e INTERNAL_ENDPOINT=$INTERNAL_ENDPOINT -e DOCKERCLOUD_REST_HOST=$DOCKERCLOUD_REST_HOST -e JWT_URL=$JWT_URL -e JWK_URL=$JWK_URL docker4x/cloud-aws:$DOCKER_FOR_IAAS_VERSION
fi

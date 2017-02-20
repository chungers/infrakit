set +e

docker node inspect self
if [ $? -ne 0 ]; then
  while :
  do
    docker swarm join {{ SPEC.SwarmJoinIP }} --token {{ SWARM_JOIN_TOKENS.Worker }}
    if [ $? -eq 0 ]; then
      $docker_run --rm \
        -e DOCKER_FOR_IAAS_VERSION \
        -e ACCOUNT_ID \
        -e REGION \
        -e CHANNEL \
        $docker_socket \
        $docker_cli \
        $guide_image \
        /usr/bin/buoy.sh "node:join"

      break
    fi

    sleep 5
  done
fi

set -e

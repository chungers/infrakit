set +e

docker node inspect self
if [ $? -ne 0 ]; then
  MANAGER="{{ SPEC.SwarmJoinIP }}"
  if [[ "$(hostname)" == "${MANAGER}" ]]; then
    MANAGER="${MANAGER/-1/-2}"
  fi

  while :
  do
    docker swarm join ${MANAGER} --token {{ SWARM_JOIN_TOKENS.Manager }}
    if [ $? -eq 0 ]; then
      $docker_run --rm \
        -e DOCKER_FOR_IAAS_VERSION \
        -e ACCOUNT_ID \
        -e REGION \
        -e CHANNEL \
        $docker_socket \
        $docker_cli \
        $guide_image \
        /usr/bin/buoy.sh "node:manager_join"

      # TEMP: remove old managers with the same name
      MANAGERS_DOWN=$(docker node ls | grep $(hostname) | awk '/-manager-/ { if ($3 == "Down" || $4 == "Down") print $1}')
      if [ -n "${MANAGERS_DOWN}" ]; then
        docker node demote ${MANAGERS_DOWN} || true
        docker node rm ${MANAGERS_DOWN} || true
      fi

      break
    fi

    sleep 5
  done
fi

set -e

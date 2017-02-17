function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

set +e

docker node inspect self
if [ $? -ne 0 ]; then
  NETWORK=$(metadata 'instance/network-interfaces/0/network' | cut -d "/" -f 4)
  STACK=${NETWORK/-network/}

  MANAGER="${STACK}-manager-1"
  if [[ "$(hostname)" == "${MANAGER}" ]]; then
    MANAGER="${STACK}-manager-2"
  fi

  while :
  do
    docker swarm join ${MANAGER} --token {{.JOIN_TOKEN}}
    if [ $? -eq 0 ]; then
      $docker_run --rm \
        -e NODE_TYPE \
        -e DOCKER_FOR_IAAS_VERSION \
        -e ACCOUNT_ID \
        -e REGION \
        -e CHANNEL \
        $docker_socket \
        $docker_cli \
        $guide_image \
        /usr/bin/buoy.sh "node:join"

      # TEMP: remove old managers with the same name
      if [ "${NODE_TYPE}" != "worker" ]; then
        MANAGERS_DOWN=$(docker node ls | grep $(hostname) | awk '/-manager-/ { if ($3 == "Down" || $4 == "Down") print $1}')
        if [ -n "${MANAGERS_DOWN}" ]; then
          echo "REMOVE"
          docker node demote ${MANAGERS_DOWN} || true
          docker node rm ${MANAGERS_DOWN} || true
        fi
      fi

      break
    fi

    sleep 5
  done
fi

set -e

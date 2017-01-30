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
      break
    fi

    sleep 5
  done
fi

set -e

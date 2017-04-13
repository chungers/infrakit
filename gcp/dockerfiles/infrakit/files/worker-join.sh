set +e

cat << EOF > /etc/docker/daemon.json
{
  "labels": {{ INFRAKIT_LABELS | to_json }}
}
EOF
kill -s HUP $(cat /var/run/docker.pid)

while ! docker info | grep -q NodeID; do
  if docker swarm join {{ SPEC.SwarmJoinIP }} --token {{ SWARM_JOIN_TOKENS.Worker }}; then
    break
  fi

  sleep 5
done

$docker_run --rm \
  -e DOCKER_FOR_IAAS_VERSION \
  -e ACCOUNT_ID \
  -e REGION \
  -e CHANNEL \
  $docker_socket \
  $docker_cli \
  $guide_image \
  /usr/bin/buoy.sh "node:join"

set -e

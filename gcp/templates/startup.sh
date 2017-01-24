#!/bin/sh

set -ex

echo This is a {{type}} node

shell_image="docker4x/shell-gcp:{{ VERSION }}"
guide_image="docker4x/guide-gcp:{{ VERSION }}"
lb_image="docker4x/l4controller-gcp:{{ VERSION }}"
infrakit_image="infrakit/devbundle:{{ VERSION }}"
infrakit_gcp_image="infrakit/gcp:{{ VERSION }}"

docker_run='docker run --label com.docker.editions.system --log-driver=json-file'
docker_daemon="$docker_run --rm -d"
docker_socket='-v /var/run/docker.sock:/var/run/docker.sock'
docker_cli='-v /usr/bin/docker:/usr/bin/docker'

function dockerPull {
  for i in $(seq 1 60); do docker pull $1 && break || sleep 1; done
}

function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

echo Start sshd

dockerPull ${shell_image}

docker inspect etc >/dev/null 2>&1 || $docker_run --name=etc -v /etc $shell_image true
$docker_run --volumes-from=etc $shell_image /usr/bin/ssh-keygen.sh

$docker_daemon --name=accounts \
  -v /dev/log:/dev/log \
  -v /home:/home \
  --volumes-from=etc \
  $shell_image \
  /usr/bin/google_accounts_daemon

$docker_daemon --name=ipforwarding \
  -v /dev/log:/dev/log \
  --cap-add=NET_ADMIN \
  --net=host \
  $shell_image \
  /usr/bin/google_ip_forwarding_daemon -d

$docker_daemon --name=shell \
  -p 22:22 \
  $docker_socket \
  $docker_cli \
  -v /var/log:/var/log \
  -v /home:/home \
  --volumes-from=etc \
  --net=host \
  $shell_image

echo Start guide

dockerPull ${guide_image}
$docker_daemon --name=guide \
  -e RUN_VACUUM="{{ properties['enableSystemPrune'] }}" \
  $docker_socket \
  $docker_cli \
  $guide_image

{% if (type in ['manager', 'leader']) %}
echo Start infrakit

local_store='-v /infrakit/:/root/.infrakit/'
run_plugin="$docker_daemon $local_store"

dockerPull ${infrakit_image}
dockerPull ${infrakit_gcp_image}
$run_plugin --name=flavor-combo $infrakit_image infrakit-flavor-combo --log=5
$run_plugin --name=flavor-swarm $docker_socket $infrakit_image infrakit-flavor-swarm --log=5
$run_plugin --name=flavor-vanilla $infrakit_image infrakit-flavor-vanilla --log=5
$run_plugin --name=group-stateless $infrakit_image infrakit-group-default --name=group-stateless --poll-interval=30s --log=5
$run_plugin --name=instance-gcp $infrakit_gcp_image infrakit-instance-gcp --log=5
$run_plugin --name=manager $docker_socket $infrakit_image infrakit-manager swarm --proxy-for-group=group-stateless --name=group --log=5

echo Start Load Balancer Listener

dockerPull ${lb_image}
$docker_daemon --name=lbcontroller $docker_socket $lb_image run --log=5
{% endif -%}

{% if (type in ['leader']) %}
echo Initialize Swarm

docker node inspect self || docker swarm init --advertise-addr eth0:2377 --listen-addr eth0:2377
{% endif -%}

{% if (type in ['manager', 'leader']) %}
echo Configure Infrakit

infrakit="$docker_run --rm $local_store $infrakit_image infrakit"
configs=/infrakit/configs
mkdir -p $configs

set +e

function watchInfrakitConfigurationChanges {
  INFRAKIT_UPDATE="2000-01-01T00:00:00.000000000Z"

  while true; do
    echo Listening for changes in Infrakit configuration $(date)

    ACCESS_TOKEN=$(metadata 'instance/service-accounts/default/token' | jq -r '.access_token')
    INFRAKIT_JSON=$(curl -s -f -X POST -d "{\"newerThan\": \"${INFRAKIT_UPDATE}\"}" -H 'Content-Type: application/json' -H "Authorization":"Bearer ${ACCESS_TOKEN}" https://runtimeconfig.googleapis.com/v1beta1/projects/{{ PROJECT }}/configs/{{ STACK }}-config/variables/infrakit:watch)
    if [ $? -ne 0 ]; then
      sleep 1
      continue
    fi

    INFRAKIT_UPDATE=$(echo "${INFRAKIT_JSON}" | jq -r '.updateTime')
    echo Updated infrakit configuration at ${INFRAKIT_UPDATE}

    IS_LEADER=$(docker node inspect self | jq -r '.[0].ManagerStatus.Leader')
    if [ "${IS_LEADER}" == "true" ]; then
      echo "${INFRAKIT_JSON}" | jq -r '.text'| jq -r '.workers' > $configs/workers.json
      echo "${INFRAKIT_JSON}" | jq -r '.text'| jq -r '.managers' > $configs/managers.json

      for i in $(seq 1 60); do $infrakit group commit /root/.infrakit/configs/workers.json && break || sleep 1; done
      for i in $(seq 1 60); do $infrakit group commit /root/.infrakit/configs/managers.json && break || sleep 1; done
    fi
  done
}
watchInfrakitConfigurationChanges &
{% endif -%}

#!/bin/sh

set -ex

echo This is a {{type}} node

DOCKER_FOR_IAAS_VERSION='{{ VERSION }}'

docker_run='docker run --label com.docker.editions.system --log-driver=json-file'
docker_daemon="$docker_run -d"
docker_socket='-v /var/run/docker.sock:/var/run/docker.sock'
docker_cli='-v /usr/bin/docker:/usr/bin/docker'

function dockerPull {
  for i in $(seq 1 60); do docker pull $1 && break || sleep 1; done
}

function dockerRm {
  docker rm -f $1 >/dev/null 2>&1 | true
}

function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

echo Start sshd

shell_image="docker4x/shell-gcp:$DOCKER_FOR_IAAS_VERSION"
dockerPull ${shell_image}

regenerateSshKeys=0
instanceId=$(metadata 'instance/id')

set +e
instanceIdOnDisk=$($docker_run \
  --volumes-from=etc \
  busybox \
  cat /etc/instance_id)
EXIT_CODE=$?
set -e

if [ $EXIT_CODE -ne 0 ]; then
  regenerateSshKeys=1
elif [ "${instanceId}" != "${instanceIdOnDisk}" ]; then
  regenerateSshKeys=1
fi

if [ ${regenerateSshKeys} -eq 1 ]; then
  dockerRm etc
  $docker_run --name=etc \
    -v /etc \
    --entrypoint ssh-keygen \
    $shell_image \
    -A

  $docker_run \
    --volumes-from=etc \
    busybox \
    /bin/sh -c "echo ${instanceId} > /etc/instance_id"
fi

dockerRm accounts
$docker_daemon --name=accounts \
  -v /dev/log:/dev/log \
  -v /home:/home \
  --volumes-from=etc \
  $shell_image \
  /usr/bin/google_accounts_daemon

dockerRm ipforwarding
$docker_daemon --name=ipforwarding \
  -v /dev/log:/dev/log \
  --cap-add=NET_ADMIN \
  --net=host \
  $shell_image \
  /usr/bin/google_ip_forwarding_daemon -d

dockerRm shell
$docker_daemon --name=shell -p 22:22 \
  $docker_socket \
  $docker_cli \
  -v /var/log:/var/log \
  -v /home:/home \
  --volumes-from=etc \
  --net=host \
  $shell_image

echo Start guide

guide_image="docker4x/guide-gcp:$DOCKER_FOR_IAAS_VERSION"
dockerPull ${guide_image}

dockerRm guide
$docker_daemon --name=guide \
  -e RUN_VACUUM="{{ properties['enableSystemPrune'] }}" \
  $docker_socket \
  $docker_cli \
  $guide_image

{% if (type in ['manager', 'leader']) %}
echo Start infrakit

local_store='-v /infrakit/:/root/.infrakit/'
run_plugin="$docker_daemon $local_store"

infrakit_image="infrakit/devbundle:$DOCKER_FOR_IAAS_VERSION"
dockerPull ${infrakit_image}

infrakit_gcp_image="infrakit/gcp:$DOCKER_FOR_IAAS_VERSION"
dockerPull ${infrakit_gcp_image}

dockerRm flavor-combo
$run_plugin --name=flavor-combo $infrakit_image infrakit-flavor-combo --log=5

dockerRm flavor-swarm
$run_plugin --name=flavor-swarm $docker_socket $infrakit_image infrakit-flavor-swarm --log=5

dockerRm flavor-vanilla
$run_plugin --name=flavor-vanilla $infrakit_image infrakit-flavor-vanilla --log=5

dockerRm group-stateless
$run_plugin --name=group-stateless $infrakit_image infrakit-group-default --name=group-stateless --log=5

dockerRm instance-gcp
$run_plugin --name=instance-gcp $infrakit_gcp_image infrakit-instance-gcp --log=5

dockerRm manager
$run_plugin --name=manager $docker_socket $infrakit_image infrakit-manager swarm --proxy-for-group=group-stateless --name=group --log=5

echo Start Load Balancer Listener

lb_image="docker4x/l4controller-gcp:$DOCKER_FOR_IAAS_VERSION"
dockerPull ${lb_image}

dockerRm lbcontroller
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

PROJECT=$(curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/project/project-id)
INFRAKIT_UPDATE="2000-01-01T00:00:00.000000000Z"

while true; do
  echo Listening for changes in Infrakit configuration $(date)

  set +e
  ACCESS_TOKEN=$(curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token | jq -r '.access_token')
  INFRAKIT_JSON=$(curl -s -f -X POST -d "{\"newerThan\": \"${INFRAKIT_UPDATE}\"}" -H 'Content-Type: application/json' -H "Authorization":"Bearer ${ACCESS_TOKEN}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/docker-config/variables/infrakit:watch)
  if [ $? -ne 0 ]; then
    sleep 1
  else
    INFRAKIT_UPDATE=$(echo "${INFRAKIT_JSON}" | jq -r '.updateTime')
    echo Updated infrakit configuration at ${INFRAKIT_UPDATE}

    IS_LEADER=$(docker node inspect self | jq -r '.[0].ManagerStatus.Leader')
    if [ "${IS_LEADER}" == "true" ]; then
      echo "${INFRAKIT_JSON}" | jq -r '.text'| jq -r '.workers' > $configs/workers.json
      echo "${INFRAKIT_JSON}" | jq -r '.text'| jq -r '.managers' > $configs/managers.json

      for i in $(seq 1 60); do $infrakit group commit /root/.infrakit/configs/workers.json && break || sleep 1; done
      for i in $(seq 1 60); do $infrakit group commit /root/.infrakit/configs/managers.json && break || sleep 1; done
    fi
  fi
  set -e
done
{% endif -%}

#!/bin/bash

set -ex

DOCKER_FOR_IAAS_VERSION='{{ VERSION }}'

shell_image="docker4x/shell-gcp:$DOCKER_FOR_IAAS_VERSION"
lb_image="docker4x/l4controller-gcp:$DOCKER_FOR_IAAS_VERSION"
infrakit_image="infrakit/devbundle:$DOCKER_FOR_IAAS_VERSION"
infrakit_gcp_image="infrakit/gcp:$DOCKER_FOR_IAAS_VERSION"

docker_run='docker run --log-driver=json-file'
docker_daemon="$docker_run -d --restart=always"
docker_client='-v /var/run/docker.sock:/var/run/docker.sock'

echo Start sshd

for i in $(seq 1 60); do docker pull $shell_image && break || sleep 1; done

docker volume create --name sshkey

$docker_run -ti --rm \
  --user root \
  -v sshkey:/etc/ssh \
  --entrypoint ssh-keygen \
  $shell_image \
  -A

$docker_run --name=shell --restart=always -d -p 22:22 \
  $docker_client \
  -v /usr/bin/docker:/usr/bin/docker \
  -v /dev/log:/dev/log \
  -v /var/log:/var/log \
  -v /home:/home \
  -v sshkey:/etc/ssh \
  $shell_image

docker exec -d shell /usr/bin/google_accounts_daemon -d

echo Start infrakit

local_store='-v /infrakit/:/root/.infrakit/'
run_plugin="$docker_daemon $local_store"

for i in $(seq 1 60); do docker pull $infrakit_image && break || sleep 1; done
for i in $(seq 1 60); do docker pull $infrakit_gcp_image && break || sleep 1; done

$run_plugin --name=flavor-combo $infrakit_image infrakit-flavor-combo --log=5
$run_plugin --name=flavor-swarm $docker_client $infrakit_image infrakit-flavor-swarm --log=5
$run_plugin --name=flavor-vanilla $infrakit_image infrakit-flavor-vanilla --log=5
$run_plugin --name=group-stateless $infrakit_image infrakit-group-default --name=group-stateless --log=5
$run_plugin --name=instance-gcp $infrakit_gcp_image infrakit-instance-gcp --log=5
$run_plugin --name=manager $docker_client $infrakit_image infrakit-manager swarm --proxy-for-group=group-stateless --name=group --log=5

echo Start Load Balancer Listener

for i in $(seq 1 60); do docker pull $lb_image && break || sleep 1; done

$docker_daemon --name=l4controller-gcp $docker_client $lb_image run --log=5

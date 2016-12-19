#!/bin/bash

set -ex

DOCKER_FOR_IAAS_VERSION="gcp-v1.13.0-rc2-beta12"

service docker start

plugins=/infrakit/plugins
mkdir -p $plugins

image="infrakit/devbundle:master-1041"
imagegcp="infrakit/gcp:master-12"
local_store="-v /infrakit/:/root/.infrakit/"
docker_client="-v /var/run/docker.sock:/var/run/docker.sock"
docker_run="docker run --log-driver=json-file -d --restart=always"
run_plugin="$docker_run $local_store"

for i in $(seq 1 60); do docker pull $image && break || sleep 1; done
for i in $(seq 1 60); do docker pull $imagegcp && break || sleep 1; done

$run_plugin --name=flavor-combo $image infrakit-flavor-combo --log=5
$run_plugin --name=flavor-swarm $docker_client $image infrakit-flavor-swarm --log=5
$run_plugin --name=flavor-vanilla $image infrakit-flavor-vanilla --log=5
$run_plugin --name=group-stateless $image infrakit-group-default --name=group-stateless --log=5
$run_plugin --name=instance-gcp $imagegcp infrakit-instance-gcp --log=5
$run_plugin --name=manager $docker_client $image infrakit-manager swarm --proxy-for-group=group-stateless --name=group --log=5

$docker_run --name=l4controller-gcp $docker_client docker4x/l4controller-gcp:$DOCKER_FOR_IAAS_VERSION run --log=5

#!/bin/bash

set -ex

service docker start

plugins=/infrakit/plugins
mkdir -p $plugins

image="infrakit/devbundle:master-1041"
imagegcp="infrakit/gcp:master-12"
local_store="-v /infrakit/:/root/.infrakit/"
docker_client="-v /var/run/docker.sock:/var/run/docker.sock"
run_plugin="docker run -d --restart always $local_store"

for i in $(seq 1 60); do docker pull $image && break || sleep 1; done
for i in $(seq 1 60); do docker pull $imagegcp && break || sleep 1; done

$run_plugin $image infrakit-flavor-combo --log 5
$run_plugin $docker_client $image infrakit-flavor-swarm --log 5
$run_plugin $image infrakit-flavor-vanilla --log 5
$run_plugin $image infrakit-group-default --name group-stateless --log 5
$run_plugin $imagegcp infrakit-instance-gcp --log 5
$run_plugin $docker_client $image infrakit-manager swarm --proxy-for-group group-stateless --name group --log 5

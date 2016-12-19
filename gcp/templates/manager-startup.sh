#!/bin/bash

set -ex

download() {
  curl -s -o /tmp/$1 https://storage.googleapis.com/docker-for-gcp-infrakit/$1
  chmod u+x /tmp/$1
}

runPlugin() {
  download ${1}
  nohup /tmp/"$@" --log 5 >/tmp/log-${1} 2>&1 &
  sleep 1
}

service docker start

runPlugin infrakit-flavor-combo
runPlugin infrakit-flavor-swarm
runPlugin infrakit-flavor-vanilla
runPlugin infrakit-group-default --name group-stateless
runPlugin infrakit-instance-gcp
runPlugin infrakit-manager swarm --proxy-for-group group-stateless --name group

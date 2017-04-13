#!/bin/sh

set -ex

function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

function startPlugins {
  echo Start Plugins

  rm -Rf /infrakit/plugins
  infrakit plugin start --config-url file:///infrakit/plugins.json --wait --exec os \
    manager \
    flavor-combo \
    flavor-swarm \
    flavor-vanilla \
    group-default \
    instance-gcp &

  echo Plugins started
}

function checkPlugins {
  [ $(ls /infrakit/plugins/*.pid | wc -l) -eq 6 ]
}

function restartPlugins {
  echo Restart Plugins

  infrakit plugin stop --all
  while :; do [ $(ls /infrakit/plugins/*.pid | wc -l) -eq 0 ] && break || sleep 1; done
  startPlugins
}

startPlugins

PROJECT=$(metadata 'project/project-id')
NETWORK=$(metadata 'instance/network-interfaces/0/network' | cut -d "/" -f 4)
STACK=${NETWORK/-network/}
INFRAKIT_UPDATE="2000-01-01T00:00:00.000000000Z"

set +e
while :; do
  echo Listening for changes in Infrakit configuration $(date)

  ACCESS_TOKEN=$(metadata 'instance/service-accounts/default/token' | jq -r '.access_token')
  INFRAKIT_JSON=$(curl -s -f -X POST -d "{\"newerThan\": \"${INFRAKIT_UPDATE}\"}" -H 'Content-Type: application/json' -H "Authorization":"Bearer ${ACCESS_TOKEN}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/${STACK}-config/variables/infrakit:watch)
  if [ $? -eq 0 ]; then
    INFRAKIT_UPDATE=$(echo "${INFRAKIT_JSON}" | jq -r '.updateTime')
    echo Updated infrakit configuration at ${INFRAKIT_UPDATE}
    echo "${INFRAKIT_JSON}" | jq -r '.text' > /infrakit/groups.json

    IS_LEADER=$(docker node inspect self | jq -r '.[0].ManagerStatus.Leader')
    if [ "${IS_LEADER}" == "true" ]; then
      checkPlugins || restartPlugins

      for i in $(seq 1 60); do
        infrakit manager commit file:///infrakit/groups.json && break || sleep 1
      done
    fi
  fi
done

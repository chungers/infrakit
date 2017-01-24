#!/bin/bash

set -ex

function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

infrakit-flavor-combo --log=5 &
infrakit-flavor-swarm --log=5 &
infrakit-flavor-vanilla --log=5 &
infrakit-group-default --name=group-stateless --poll-interval=30s --log=5 &
sleep 1
infrakit-instance-gcp --log=5 &
infrakit-manager swarm --proxy-for-group=group-stateless --name=group --log=5 &

PROJECT=$(metadata 'project/project-id')
NETWORK=$(metadata 'instance/network-interfaces/0/network' | cut -d "/" -f 4)
STACK=${NETWORK/-network/}
INFRAKIT_UPDATE="2000-01-01T00:00:00.000000000Z"

set +e
while true; do
  echo Listening for changes in Infrakit configuration $(date)

  ACCESS_TOKEN=$(metadata 'instance/service-accounts/default/token' | jq -r '.access_token')
  INFRAKIT_JSON=$(curl -s -f -X POST -d "{\"newerThan\": \"${INFRAKIT_UPDATE}\"}" -H 'Content-Type: application/json' -H "Authorization":"Bearer ${ACCESS_TOKEN}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/${STACK}-config/variables/infrakit:watch)
  if [ $? -ne 0 ]; then
    sleep 1
    continue
  fi

  INFRAKIT_UPDATE=$(echo "${INFRAKIT_JSON}" | jq -r '.updateTime')
  echo Updated infrakit configuration at ${INFRAKIT_UPDATE}

  IS_LEADER=$(docker node inspect self | jq -r '.[0].ManagerStatus.Leader')
  if [ "${IS_LEADER}" == "true" ]; then
    echo "${INFRAKIT_JSON}" | jq -r '.text'| jq -r '.workers' > /workers.json
    echo "${INFRAKIT_JSON}" | jq -r '.text'| jq -r '.managers' > /managers.json

    for i in $(seq 1 60); do infrakit group commit /workers.json && break || sleep 1; done
    for i in $(seq 1 60); do infrakit group commit /managers.json && break || sleep 1; done
  fi
done

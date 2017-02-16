#!/bin/sh

set -ex

function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

PROJECT=$(metadata 'project/project-id')
NETWORK=$(metadata 'instance/network-interfaces/0/network' | cut -d "/" -f 4)
STACK=${NETWORK/-network/}
INFRAKIT_UPDATE="2000-01-01T00:00:00.000000000Z"

infrakit-manager --name=group --proxy-for-group=group-stateless swarm --log=5 &
infrakit plugin start --wait --config-url file:///infrakit/plugins.json --os flavor-combo flavor-swarm flavor-vanilla group-default instance-gcp &
sleep 5  # manager needs to detect leadership

set +e
while :
do
  echo Listening for changes in Infrakit configuration $(date)

  ACCESS_TOKEN=$(metadata 'instance/service-accounts/default/token' | jq -r '.access_token')
  INFRAKIT_JSON=$(curl -s -f -X POST -d "{\"newerThan\": \"${INFRAKIT_UPDATE}\"}" -H 'Content-Type: application/json' -H "Authorization":"Bearer ${ACCESS_TOKEN}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/${STACK}-config/variables/infrakit:watch)
  if [ $? -ne 0 ]; then
    sleep 5
    continue
  fi

  INFRAKIT_UPDATE=$(echo "${INFRAKIT_JSON}" | jq -r '.updateTime')
  echo Updated infrakit configuration at ${INFRAKIT_UPDATE}

  IS_LEADER=$(docker node inspect self | jq -r '.[0].ManagerStatus.Leader')
  if [ "${IS_LEADER}" == "true" ]; then
    echo "${INFRAKIT_JSON}" | jq -r '.text' > /infrakit/groups.json

    for i in $(seq 1 60); do infrakit manager commit file:///infrakit/groups.json && infrakit group ls && break || sleep 1; done
  fi
done

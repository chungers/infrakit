#!/bin/sh

set -ex

function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

rm -Rf ~/.infrakit
infrakit plugin start --config-url file:///infrakit/plugins.json --exec os \
  manager \
  flavor-combo \
  flavor-swarm \
  flavor-vanilla \
  group-default \
  instance-gcp &

PROJECT=$(metadata 'project/project-id')
NETWORK=$(metadata 'instance/network-interfaces/0/network' | cut -d "/" -f 4)
STACK=${NETWORK/-network/}
INFRAKIT_UPDATE="2000-01-01T00:00:00.000000000Z"

echo "Plugins started."

set +e
while :
do
  echo Listening for changes in Infrakit configuration $(date)

  ACCESS_TOKEN=$(metadata 'instance/service-accounts/default/token' | jq -r '.access_token')
  INFRAKIT_JSON=$(curl -s -f -m 10 -X POST -d "{\"newerThan\": \"${INFRAKIT_UPDATE}\"}" -H 'Content-Type: application/json' -H "Authorization":"Bearer ${ACCESS_TOKEN}" https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/${STACK}-config/variables/infrakit:watch)
  if [ $? -eq 0 ]; then
    INFRAKIT_UPDATE=$(echo "${INFRAKIT_JSON}" | jq -r '.updateTime')
    echo Updated infrakit configuration at ${INFRAKIT_UPDATE}
    echo "${INFRAKIT_JSON}" | jq -r '.text' > /infrakit/groups.json
  fi

  if [ -f /infrakit/groups.json ]; then
    IS_LEADER=$(docker node inspect self | jq -r '.[0].ManagerStatus.Leader')
    if [ "${IS_LEADER}" == "true" ]; then
      for i in $(seq 1 60); do infrakit manager commit file:///infrakit/groups.json && break || sleep 1; done
    fi
  fi
done

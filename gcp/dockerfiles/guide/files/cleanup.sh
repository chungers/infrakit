#!/bin/sh

set -e

echo "Cleaning up nodes..."

# this script cleans up and nodes that have been upgraded and no longer need to be in the swarm.
if [ "$NODE_TYPE" == "worker" ] ; then
  # this doesn't run on workers, only leader
  exit 0
fi

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')
if [ "$IS_LEADER" != "true" ]; then
  # this doesn't run on workers, only leader
  exit 0
fi

function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

# Find the list of workers in the stack
PROJECT=$(metadata project/project-id)
ZONE=$(metadata instance/zone | awk -F/ '{print $NF}')
NETWORK=$(metadata 'instance/network-interfaces/0/network' | cut -d "/" -f 4)
STACK=${NETWORK/-network/}
AUTH=$(metadata instance/service-accounts/default/token | jq -r ".access_token")
WORKERS=$(curl -s "https://www.googleapis.com/compute/v1/projects/${PROJECT}/zones/${ZONE}/instances/?filter=name+eq+${STACK}-worker-%5Cw*" -H "Authorization: Bearer ${AUTH}" | jq -r '.items[].name')

# Remove a worker if it's 'Down' and not anymore in the stack
WORKERS_DOWN=$(docker node ls | awk '/-worker-/ { if ($3 == "Down") print $2}')
for WORKER_DOWN in ${WORKERS_DOWN}; do
  for WORKER in ${WORKERS}; do
    if [ "${WORKER}" == "${WORKER_DOWN}" ]; then
      continue 2
    fi
  done

  docker node rm ${WORKER_DOWN} || true
done

# Remove a manager that is 'Down' and has the same name as a 'Ready' manager
MANAGERS_READY=$(docker node ls | awk '/-manager-/ { if ($3 == "Ready") print $2; if ($4 == "Ready") print $3}')
for MANAGER_READY in ${MANAGERS_READY}; do
  MANAGERS_DOWN=$(docker node ls | grep $MANAGER_READY | awk '/-manager-/ { if ($3 == "Down" || $4 == "Down") print $1}')
  if [ -n "${MANAGERS_DOWN}" ]; then
    docker node demote ${MANAGERS_DOWN} || true
    docker node rm ${MANAGERS_DOWN} || true
  fi
done

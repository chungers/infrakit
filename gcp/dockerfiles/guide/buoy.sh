#!/bin/bash

if [[ "$1" == "node:join" ]]; then
  if [[ "$NODE_TYPE" == "worker" ]] ; then
    NODE_ID=$(docker info | grep NodeID | cut -f2 -d: | sed -e 's/^[ \t]*//')
    SWARM_ID='n/a' #TODO:FIX add this for workers.

    /usr/bin/buoy -event="node:join" \
      -swarm_id=$SWARM_ID \
      -node_id=$NODE_ID \
      -channel=$CHANNEL
  else
    NODE_ID=$(docker node inspect self | jq -r '.[].ID')
    SWARM_ID=$(docker info | grep ClusterID | cut -f2 -d: | sed -e 's/^[ \t]*//')

    /usr/bin/buoy -event="node:manager_join" \
      -swarm_id=$SWARM_ID \
      -node_id=$NODE_ID \
      -channel=$CHANNEL
  fi

  exit 0
fi

if [[ "$NODE_TYPE" == "worker" ]] ; then
  # this doesn't run on workers, only managers.
  exit 0
fi

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')
if [[ "$IS_LEADER" == "true" ]]; then
  if [[ "$1" == "identify" ]]; then
    /usr/bin/buoy -event="identify" -iaas_provider="gcp"
    exit 0
  fi

  SWARM_ID=$(docker info | grep ClusterID | cut -f2 -d: | sed -e 's/^[ \t]*//')
  NODE_ID=$(docker node inspect self | jq -r '.[].ID')
  DOCKER_VERSION=$(docker version --format '{{.Server.Version}}')

  if [[ "$1" == "swarm:init" ]]; then
    /usr/bin/buoy -event="swarm:init" \
      -swarm_id=$SWARM_ID \
      -node_id=$NODE_ID \
      -channel=$CHANNEL \
      -docker_version=$DOCKER_VERSION
    exit 0
  fi

  NUM_MANAGERS=$(docker info | grep Managers | cut -f2 -d: | sed -e 's/^[ \t]*//')
  TOTAL_NODES=$(docker info | grep Nodes | cut -f2 -d: | sed -e 's/^[ \t]*//')
  NUM_WORKERS=$(expr $TOTAL_NODES - $NUM_MANAGERS)
  NUM_SERVICES=$(docker service ls -q | wc -w)

  /usr/bin/buoy -event="swarm:ping" \
    -swarm_id=$SWARM_ID \
    -node_id=$NODE_ID \
    -channel=$CHANNEL \
    -docker_version=$DOCKER_VERSION \
    -managers=$NUM_MANAGERS \
    -workers=$NUM_WORKERS \
    -services=$NUM_SERVICES
fi

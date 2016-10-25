#!/bin/bash
# this script refreshes the swarm tokens in azure table if they have changed.
if [ "$NODE_TYPE" == "worker" ] ; then
    # this doesn't run on workers, only managers.
    exit 0
fi

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
    # we are the leader, We only need to call once, so we only call from the current leader.
    DATA=$(python /usr/bin/azuretokens.py get-tokens)
    MANAGER_IP=$(echo $DATA | cut -d'|' -f 1)
    STORED_MANAGER_TOKEN=$(echo $DATA | cut -d'|' -f 2)
    STORED_WORKER_TOKEN=$(echo $DATA | cut -d'|' -f 3)

    MANAGER_TOKEN=$(docker swarm join-token manager -q)
    WORKER_TOKEN=$(docker swarm join-token worker -q)

    if [[ "$STORED_MANAGER_TOKEN" != "$MANAGER_TOKEN" ]] || [[ "$STORED_WORKER_TOKEN" != "$WORKER_TOKEN" ]]; then
        echo "Swarm tokens changed, updating azure table with new tokens"
        python /usr/bin/azuretokens.py insert-tokens $MANAGER_IP $MANAGER_TOKEN $WORKER_TOKEN
    fi

fi

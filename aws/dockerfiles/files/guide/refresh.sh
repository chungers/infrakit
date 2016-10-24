#!/bin/bash
# this script refreshes the swarm tokens in dynamodb if they have changed.
if [ "$NODE_TYPE" == "worker" ] ; then
    # this doesn't run on workers, only managers.
    exit 0
fi

# make sure we are not in process of shutting down.
if [ -e /tmp/.shutdown-init ]
then
    echo "We are shutting down, no need to continue."
    # shutdown has initialized, don't start because we might not be able to finish.
    exit 0
fi

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
    # we are the leader, We only need to call once, so we only call from the current leader.
    MANAGER=$(aws dynamodb get-item --region $REGION --table-name $DYNAMODB_TABLE --key '{"node_type":{"S": "primary_manager"}}')
    MANAGER_IP=$(echo $MANAGER | jq -r '.Item.ip.S')
    STORED_MANAGER_TOKEN=$(echo $MANAGER | jq -r '.Item.manager_token.S')
    STORED_WORKER_TOKEN=$(echo $MANAGER | jq -r '.Item.worker_token.S')

    MANAGER_TOKEN=$(docker swarm join-token manager -q)
    WORKER_TOKEN=$(docker swarm join-token worker -q)

    if [[ "$STORED_MANAGER_TOKEN" != "$MANAGER_TOKEN" ]] || [[ "$STORED_WORKER_TOKEN" != "$WORKER_TOKEN" ]]; then
        echo "Swarm tokens changed, updating dynamodb with new tokens"
        aws dynamodb update-item \
            --table-name $DYNAMODB_TABLE \
            --region $REGION \
            --key '{"node_type":{"S": "primary_manager"}}' \
            --update-expression 'SET manager_token=:m, worker_token=:w' \
            --expression-attribute-values '{":m": {"S":"'"$MANAGER_TOKEN"'"}, ":w": {"S":"'"$WORKER_TOKEN"'"}}' \
            --return-consumed-capacity TOTAL
    fi

fi

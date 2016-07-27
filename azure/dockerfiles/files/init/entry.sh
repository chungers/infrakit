#!/bin/bash
echo "#================"
echo "Start Swarm setup"

echo "PATH=$PATH"
echo "ROLE=$ROLE"
echo "MANAGER_IP=$MANAGER_IP"
echo "PRIVATE_IP=$PRIVATE_IP"
echo "DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION"
echo "ACCOUNT_ID=$ACCOUNT_ID"
echo "REGION=$REGION"
echo "#================"

get_swarm_id()
{
    if [ "$ROLE" == "MANAGER" ] ; then
        export SWARM_ID=$(docker info | grep ClusterID | cut -f2 -d: | sed -e 's/^[ \t]*//')
    else
        # not available in docker info. might be available in future release.
        export SWARM_ID='n/a'
    fi
    echo "SWARM_ID: $SWARM_ID"
}

get_node_id()
{
    export NODE_ID=$(docker info | grep NodeID | cut -f2 -d: | sed -e 's/^[ \t]*//')
    echo "NODE: $NODE_ID"
}

get_tokens()
{
    export MANAGER_TOKEN=$(docker -H $MANAGER_IP:2375 swarm join-token manager -q)
    export WORKER_TOKEN=$(docker -H $MANAGER_IP:2375 swarm join-token worker -q)
    echo "MANAGER_TOKEN=$MANAGER_TOKEN"
    echo "WORKER_TOKEN=$WORKER_TOKEN"
}

setup_manager()
{
    # we are the primary, so init the cluster
    docker swarm init --listen-addr $PRIVATE_IP:2377 --advertise-addr $PRIVATE_IP:2377

    get_swarm_id
    get_node_id

    echo "   Primary Manager init complete"
    # send identify message
    buoy -event=identify -swarm_id=$SWARM_ID -flavor=azure -node_id=$NODE_ID
}

setup_worker()
{
    echo " Setup Worker"
    echo "   MANAGER_IP=$MANAGER_IP"
    get_tokens
    docker swarm join --token $WORKER_TOKEN $MANAGER_IP:2377
    get_swarm_id
    get_node_id
    buoy -event="node:join" -swarm_id=$SWARM_ID -flavor=azure -node_id=$NODE_ID
}

# if it is a manager, setup as manager, if not, setup as worker node.
if [ "$ROLE" == "MANAGER" ] ; then
    echo " It's a Manager, run setup"
    setup_manager
else
    echo " It's a worker Node, run setup"
    setup_worker
fi

# show the results.
echo "#================ docker info    ==="
docker info
echo "#================ docker node ls ==="
docker node ls
echo "#==================================="
echo "Complete Swarm setup"

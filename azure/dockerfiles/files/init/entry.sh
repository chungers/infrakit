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
    echo "Get MANAGER IP from Azure Table"
    export MANAGER_IP=$(python azureleader.py get-ip)
    get_manager_token
    get_worker_token
}

get_manager_token()
{
    if [ -n "$MANAGER_IP" ]; then
        export MANAGER_TOKEN=$(wget -qO- http://$MANAGER_IP:9024/token/manager/)
        echo "MANAGER_TOKEN=$MANAGER_TOKEN"
    else
        echo "MANAGER_TOKEN can't be found yet. MANAGER_IP isn't set yet."
    fi
}

get_worker_token()
{
    if [ -n "$MANAGER_IP" ]; then
        export WORKER_TOKEN=$(wget -qO- http://$MANAGER_IP:9024/token/worker/)
        echo "WORKER_TOKEN=$WORKER_TOKEN"
    else
        echo "WORKER_TOKEN can't be found yet. MANAGER_IP isn't set yet."
    fi
}

confirm_primary_ready()
{
    n=0
    until [ $n -ge 5 ]
    do
        get_tokens
        echo "PRIMARY_MANAGER_IP=$MANAGER_IP"
        # if Manager IP or manager_token is empty or manager_token is null, not ready yet.
        # token would be null for a short time between swarm init, and the time the
        # token is added to azure table
        if [ -z "$MANAGER_IP" ] || [ -z "$MANAGER_TOKEN" ] || [ "$MANAGER_TOKEN" == "null" ]; then
            echo "Primary manager Not ready yet, sleep for 60 seconds."
            sleep 60
            n=$[$n+1]
        else
            echo "Primary manager is ready."
            break
        fi
    done
}

join_as_secondary_manager()
{
    echo "   Secondary Manager"
    if [ -z "$MANAGER_IP" ] || [ -z "$MANAGER_TOKEN" ] || [ "$MANAGER_TOKEN" == "null" ]; then
        confirm_primary_ready
    fi
    echo "   MANAGER_IP=$MANAGER_IP"
    echo "   MANAGER_TOKEN=$MANAGER_TOKEN"
    # sleep for 30 seconds to make sure the primary manager has enough time to setup before
    # we try and join.

    sleep 30
    # we are not primary, so join as secondary manager.
    n=0
    until [ $n -ge 5 ]
    do
        docker swarm join --token $MANAGER_TOKEN --listen-addr $PRIVATE_IP:2377 --advertise-addr $PRIVATE_IP:2377 $MANAGER_IP:2377

        get_swarm_id
        get_node_id

        # check if we have a NODE_ID, if so, we were able to join, if not, it failed.
        if [ -z "$NODE_ID" ]; then
            echo "Can't connect to primary manager, sleep and try again"
            sleep 60
            n=$[$n+1]

            # query azure table again, incase the manager changed
            get_tokens
        else
            echo "Connected to primary manager, NODE_ID=$NODE_ID , SWARM_ID=$SWARM_ID"
            break
        fi
    done
    buoy -event="node:manager_join" -swarm_id=$SWARM_ID -flavor=azure -node_id=$NODE_ID
    echo "   Secondary Manager complete"
}

setup_manager()
{
    echo "Setup Manager"
    echo "   PRIVATE_IP=$PRIVATE_IP"
    echo "   PRIMARY_MANAGER_IP=$MANAGER_IP"

    if [ -z "$MANAGER_IP" ]; then
        echo "Primary Manager IP not set yet, lets try and set it."
        # try to create the azure table that will store tokens, if it succeeds then it is the first
        # and it is the primary manager. If it fails, then it isn't first, and treat the record
        # that is there, as the primary manager, and join that swarm.
        python azureleader.py create-table
        PRIMARY_RESULT=$?
        echo "   PRIMARY_RESULT=$PRIMARY_RESULT"

        if [ $PRIMARY_RESULT -eq 0 ]; then
            echo "   Primary Manager init"
            # we are the primary, so init the cluster
            docker swarm init --listen-addr $PRIVATE_IP:2377 --advertise-addr $PRIVATE_IP:2377
            # we can now get the swarm id and node id.
            get_swarm_id
            get_node_id

            # update azure table with the ip
            python azureleader.py insert-ip $PRIVATE_IP

            echo "   Primary Manager init complete"
            # send identify message
            buoy -event=identify -swarm_id=$SWARM_ID -flavor=azure -node_id=$NODE_ID
        else
            echo " Error is normal, it is because we already have a primary node, lets setup a secondary manager instead."
            join_as_secondary_manager
        fi
    elif [ "$PRIVATE_IP" == "$MANAGER_IP" ]; then
        echo "   PRIVATE_IP == MANAGER_IP, we are already the leader, maybe it was a reboot?"
        SWARM_STATE=$(docker info | grep Swarm | cut -f2 -d: | sed -e 's/^[ \t]*//')
        # should be active, pending or inactive
        echo "   Swarm State = $SWARM_STATE"
        # check if swarm is good?
    else
        echo "   join as Secondary Manager"
        join_as_secondary_manager
    fi

}

setup_worker()
{
    echo " Setup Worker"
    if [ -z "$MANAGER_IP" ] || [ -z "$WORKER_TOKEN" ] || [ "$MANAGER_TOKEN" == "null" ]; then
        confirm_primary_ready
    fi

    echo "   MANAGER_IP=$MANAGER_IP"
    # try an connect to the swarm manager.
    n=0
    until [ $n -ge 5 ]
    do
        docker swarm join --token $WORKER_TOKEN $MANAGER_IP:2377
        get_swarm_id
        get_node_id

        # check if we have a NODE_ID, if so, we were able to join, if not, it failed.
        if [ -z "$NODE_ID" ]; then
            echo "Can't connect to primary manager, sleep and try again"
            sleep 60
            n=$[$n+1]

            # query azure table again, incase the manager changed
            get_tokens
        else
            echo "Connected to manager, NODE_ID=$NODE_ID , SWARM_ID=$SWARM_ID"
            break
        fi
    done
    buoy -event="node:join" -swarm_id=$SWARM_ID -flavor=azure -node_id=$NODE_ID
}

# init variables based on azure token table contents (if populated)
get_tokens

# if it is a manager, setup as manager, if not, setup as worker node.
if [ "$ROLE" == "MANAGER" ] ; then
    echo " It's a Manager, run setup"
    get_manager_token
    setup_manager
else
    echo " It's a worker Node, run setup"
    get_worker_token
    setup_worker
fi

# show the results.
echo "#================ docker info    ==="
docker info
echo "#================ docker node ls ==="
docker node ls
echo "#==================================="
echo "Complete Swarm setup"

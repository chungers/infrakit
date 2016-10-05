#!/bin/bash
echo "#================"
echo "Start Swarm setup"

# Setup path with the docker binaries
export MYHOST=`wget -qO- http://169.254.169.254/latest/meta-data/hostname`
SWARM_STATE=$(docker info | grep Swarm | cut -f2 -d: | sed -e 's/^[ \t]*//')

echo "PATH=$PATH"
echo "NODE_TYPE=$NODE_TYPE"
echo "DYNAMODB_TABLE=$DYNAMODB_TABLE"
echo "HOSTNAME=$MYHOST"
echo "STACK_NAME=$STACK_NAME"
echo "INSTANCE_NAME=$INSTANCE_NAME"
echo "AWS_REGION=$REGION"
echo "MANAGER_IP=$MANAGER_IP"
echo "SWARM_STATE=$SWARM_STATE"
echo "#================"

get_swarm_id()
{
    if [ "$NODE_TYPE" == "manager" ] ; then
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

get_primary_manager_ip()
{
    echo "Get Primary Manager IP"
    # query dynamodb and get the Ip for the primary manager.
    MANAGER=$(aws dynamodb get-item --region $REGION --table-name $DYNAMODB_TABLE --key '{"node_type":{"S": "primary_manager"}}')
    export MANAGER_IP=$(echo $MANAGER | jq -r '.Item.ip.S')
    export MANAGER_TOKEN=$(echo $MANAGER | jq -r '.Item.manager_token.S')
    export WORKER_TOKEN=$(echo $MANAGER | jq -r '.Item.worker_token.S')
    echo "MANAGER_TOKEN=$MANAGER_TOKEN"
    echo "WORKER_TOKEN=$WORKER_TOKEN"
}

get_tokens()
{
    export MANAGER_TOKEN=$(docker swarm join-token manager -q)
    export WORKER_TOKEN=$(docker swarm join-token worker -q)
    echo "MANAGER_TOKEN=$MANAGER_TOKEN"
    echo "WORKER_TOKEN=$WORKER_TOKEN"
}

confirm_primary_ready()
{
    n=0
    until [ $n -ge 5 ]
    do
        get_primary_manager_ip
        echo "PRIMARY_MANAGER_IP=$MANAGER_IP"
        # if Manager IP or manager_token is empty or manager_token is null, not ready yet.
        # token would be null for a short time between swarm init, and the time the
        # token is added to dynamodb
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
    # we are not, join as secondary manager.
    docker swarm join --token $MANAGER_TOKEN --listen-addr $PRIVATE_IP:2377 --advertise-addr $PRIVATE_IP:2377 $MANAGER_IP:2377

    get_swarm_id
    get_node_id
    buoy -event="node:manager_join" -swarm_id=$SWARM_ID -flavor=aws -node_id=$NODE_ID
    echo "   Secondary Manager complete"
}

setup_manager()
{
    echo "Setup Manager"
    export PRIVATE_IP=`wget -qO- http://169.254.169.254/latest/meta-data/local-ipv4`

    echo "   PRIVATE_IP=$PRIVATE_IP"
    echo "   PRIMARY_MANAGER_IP=$MANAGER_IP"
    if [ -z "$MANAGER_IP" ]; then
        echo "Primary Manager IP not set yet, lets try and set it."
        # try to write to the table as the primary_manager, if it succeeds then it is the first
        # and it is the primary manager. If it fails, then it isn't first, and treat the record
        # that is there, as the primary manager, and join that swarm.
        aws dynamodb put-item \
            --table-name $DYNAMODB_TABLE \
            --region $REGION \
            --item '{"node_type":{"S": "primary_manager"},"ip": {"S":"'"$PRIVATE_IP"'"}}' \
            --condition-expression 'attribute_not_exists(node_type)' \
            --return-consumed-capacity TOTAL
        PRIMARY_RESULT=$?
        echo "   PRIMARY_RESULT=$PRIMARY_RESULT"

        if [ $PRIMARY_RESULT -eq 0 ]; then
            echo "   Primary Manager init"
            # we are the primary, so init the cluster
            docker swarm init --listen-addr $PRIVATE_IP:2377 --advertise-addr $PRIVATE_IP:2377
            # we can now get the tokens.
            get_tokens
            get_swarm_id
            get_node_id

            # update dynamodb with the tokens
            aws dynamodb put-item \
                --table-name $DYNAMODB_TABLE \
                --region $REGION \
                --item '{"node_type":{"S": "primary_manager"},"ip": {"S":"'"$PRIVATE_IP"'"},"manager_token": {"S":"'"$MANAGER_TOKEN"'"},"worker_token": {"S":"'"$WORKER_TOKEN"'"}}' \
                --return-consumed-capacity TOTAL

            echo "   Primary Manager init complete"
            # send identify message
            buoy -event=identify -swarm_id=$SWARM_ID -flavor=aws -node_id=$NODE_ID
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

setup_node()
{
    echo " Setup Node"
    # setup the node, by joining the swarm.
    if [ -z "$MANAGER_IP" ] || [ -z "$WORKER_TOKEN" ] || [ "$MANAGER_TOKEN" == "null" ]; then
        confirm_primary_ready
    fi
    echo "   MANAGER_IP=$MANAGER_IP"
    docker swarm join --token $WORKER_TOKEN $MANAGER_IP:2377
    get_swarm_id
    get_node_id
    buoy -event="node:join" -swarm_id=$SWARM_ID -flavor=aws -node_id=$NODE_ID
}

# see if the primary manager IP is already set.
get_primary_manager_ip

# if it is a manager, setup as manager, if not, setup as worker node.
if [ "$NODE_TYPE" == "manager" ] ; then
    echo " It's a Manager, run setup"
    setup_manager
else
    echo " It's a worker Node, run setup"
    setup_node
fi

# show the results.
echo "#================ docker info    ==="
docker info
echo "#================ docker node ls ==="
docker node ls
echo "#==================================="
echo "Notify AWS that server is ready"
cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION

echo "Complete Swarm setup"

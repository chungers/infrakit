#!/bin/bash
echo "#================"
echo "Start Swarm setup"

# Setup path with the docker binaries
export MYHOST=`wget -qO- http://169.254.169.254/latest/meta-data/hostname`

echo "PATH=$PATH"
echo "NODE_TYPE=$NODE_TYPE"
echo "DYNAMODB_TABLE=$DYNAMODB_TABLE"
echo "HOSTNAME=$MYHOST"
echo "STACK_NAME=$STACK_NAME"
echo "INSTANCE_NAME=$INSTANCE_NAME"
echo "AWS_REGION=$REGION"
echo "MANAGER_IP=$MANAGER_IP"
echo "#================"

get_swarm_id()
{
    if [ "$NODE_TYPE" == "manager" ] ; then
        export SWARM_ID=$(docker swarm inspect -f '{{.ID}}')
    else
        # not available in docker info. might be available in future release.
        export SWARM_ID='n/a'
    fi
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
    export MANAGER_IP=$(aws dynamodb get-item --table-name $DYNAMODB_TABLE --key '{"node_type":{"S": "primary_manager"}}' | jq -r '.Item.ip.S')
}

confirm_primary_ready()
{
    n=0
    until [ $n -ge 5 ]
    do
        get_primary_manager_ip
        echo "PRIMARY_MANAGER_IP=$MANAGER_IP"
        if [ -z "$MANAGER_IP" ]; then
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
    if [ -z "$MANAGER_IP" ]; then
        confirm_primary_ready
    fi
    echo "   MANAGER_IP=$MANAGER_IP"
    # sleep for 30 seconds to make sure the primary manager has enough time to setup before
    # we try and join.
    sleep 30
    # we are not, join as secondary manager.
    docker swarm join --manager --listen-addr $PRIVATE_IP:2377 $MANAGER_IP:2377
    docker swarm update --auto-accept manager --auto-accept worker
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
            --item '{"node_type":{"S": "primary_manager"},"ip": {"S":"'"$PRIVATE_IP"'"}}' \
            --condition-expression 'attribute_not_exists(node_type)' \
            --return-consumed-capacity TOTAL
        PRIMARY_RESULT=$?
        echo "   PRIMARY_RESULT=$PRIMARY_RESULT"

        if [ $PRIMARY_RESULT -eq 0 ]; then
            echo "   Primary Manager init"
            # we are the primary, so init the cluster
            docker swarm init --secret "" --auto-accept manager --auto-accept worker --listen-addr $PRIVATE_IP:2377
            echo "   Primary Manager init complete"
            # send identify message
            buoy -event=identify -swarm_id=$SWARM_ID -flavor=aws
        else
            echo " Error is normal, it is because we already have a primary node, lets setup a secondary manager instead."
            join_as_secondary_manager
        fi
    else
        echo "   Secondary Manager"
        join_as_secondary_manager
    fi
}

setup_node()
{
    echo " Setup Node"
    # setup the node, by joining the swarm.
    if [ -z "$MANAGER_IP" ]; then
        confirm_primary_ready
    fi
    echo "   MANAGER_IP=$MANAGER_IP"
    docker swarm join $MANAGER_IP:2377
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

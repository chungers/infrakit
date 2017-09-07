#!/bin/bash
echo "#================"
echo "Start DTR setup"

# default to yes if INSTALL DDC is empty.
INSTALL_DDC=${INSTALL_DDC:-"yes"}

DTR_TAG=${DTR_TAG-"2.2.3"}
DTR_ORG=${DTR_ORG:-"docker"}
DTR_IMAGE=${DTR_ORG}/dtr:${DTR_TAG}
DTR_HTTPS_PORT=12391
DTR_HTTP_PORT=12392
IMAGE_LIST_ARGS=''
MYIP=$(wget -qO- http://169.254.169.254/latest/meta-data/local-ipv4)
LOCAL_HOSTNAME=$(wget -qO- http://169.254.169.254/latest/meta-data/local-hostname)
DTR_DYNAMO_FIELD='dtr_replicas'

# Normalize the ELB hostnames to be all lowercase.
APP_ELB_HOSTNAME=$(tr '[:upper:]' '[:lower:]' <<< "$APP_ELB_HOSTNAME")
UCP_ELB_HOSTNAME=$(tr '[:upper:]' '[:lower:]' <<< "$UCP_ELB_HOSTNAME")
DTR_ELB_HOSTNAME=$(tr '[:upper:]' '[:lower:]' <<< "$DTR_ELB_HOSTNAME")

echo "Get Primary Manager IP"
# query dynamodb and get the Ip for the primary manager.
PRIMARY_MANAGER=$(aws dynamodb get-item --region $REGION --table-name $DYNAMODB_TABLE --key '{"node_type":{"S": "primary_manager"}}')
export PRIMARY_MANAGER_IP=$(echo $PRIMARY_MANAGER | jq -r '.Item.ip.S')

echo "PRIMARY_MANAGER_IP=$PRIMARY_MANAGER_IP"
echo "MYIP=$MYIP"
echo "LOCAL_HOSTNAME=$LOCAL_HOSTNAME"
echo "PATH=$PATH"
echo "STACK_NAME=$STACK_NAME"
echo "REGION=$REGION"
echo "S3_BUCKET_NAME=$S3_BUCKET_NAME"
echo "LICENSE=$LICENSE"
echo "INSTALL_DDC=$INSTALL_DDC"
echo "APP_ELB_HOSTNAME=$APP_ELB_HOSTNAME"
echo "UCP_ELB_HOSTNAME=$UCP_ELB_HOSTNAME"
echo "DTR_ELB_HOSTNAME=$DTR_ELB_HOSTNAME"
echo "NODE_NAME=$NODE_NAME"
echo "UCP_ADMIN_USER=$UCP_ADMIN_USER"
echo "DTR_IMAGE=$DTR_IMAGE"
echo "DTR_HTTPS_PORT=$DTR_HTTPS_PORT"
echo "DTR_HTTP_PORT=$DTR_HTTP_PORT"
echo "MANAGER_COUNT=$MANAGER_COUNT"
echo "#================"

# we don't want to install, exit now.
if [[ "$INSTALL_DDC" != "yes" ]] ; then
    exit 0
fi

# Login if credentials were provided
if [[ "$REGISTRY_PASSWORD" != "" ]] ; then
    docker login -u "${REGISTRY_USERNAME}" -p "${REGISTRY_PASSWORD}"
fi

images=$(docker run --rm $DTR_IMAGE images)
for im in $images; do
    docker pull $im
done

# ^^ If this is an upgrade, do we still need to pull down these images? Should they instead be the version that installed in the swarm?

# Checking if DTR is up
checkDTR(){
    echo "Checking to see if DTR is up and healthy"
    n=0
    until [ $n -gt 20 ];
    do
        if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$MYIP:$DTR_HTTPS_PORT/health) -eq 200 ]];
            then echo "Main DTR Replica is up! Starting DTR replica join process"
            break
        else
            if [[ $n -eq 20 ]];
                then echo "DTR failed status check after $n tries. Aborting Installation..."
                exit 1
            fi
            echo "Try #$n: checking DTR status..."
            sleep 30
            let n+=1
        fi
    done
}

echo "lets install DTR now."
echo "MYIP=$MYIP"
echo "MYIP=$DYNAMODB_TABLE"
echo "MYIP=$REGION"

# lets see if we are the first DTR node to start.
aws dynamodb put-item \
    --table-name $DYNAMODB_TABLE \
    --region $REGION \
    --item '{"node_type":{"S": "primary_dtr_manager"},"ip": {"S":"'"$MYIP"'"}}' \
    --condition-expression 'attribute_not_exists(node_type)' \
    --return-consumed-capacity TOTAL
PRIMARY_RESULT=$?
echo "   PRIMARY_RESULT=$PRIMARY_RESULT"

if [ $PRIMARY_RESULT -eq 0 ]; then
    echo "  First DTR node"
    # we are the first DTR node, so lets install DTR.

    REPLICA_ID=$(echo $MYIP | sed "s/\./0/g" | awk '{print "0000"$1}' | tail -c 13)
    echo "REPLICA_ID=$REPLICA_ID "
    DTR_LEADER_INSTALL="yes"
    echo "Installing First DTR Replica..."
    sleep 30
    echo "Install DTR"
    date
    docker run --label com.docker.editions.system  --rm "$DTR_IMAGE" install --replica-https-port "$DTR_HTTPS_PORT" --replica-http-port "$DTR_HTTP_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$NODE_NAME" --dtr-external-url $DTR_ELB_HOSTNAME:443 --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --replica-id $REPLICA_ID
    echo "After running install via Docker"
    date

    # add replica id to dynamodb so other nodes can get it.
    aws dynamodb put-item \
        --table-name $DYNAMODB_TABLE \
        --region $REGION \
        --item '{"node_type":{"S":  "'"$DTR_DYNAMO_FIELD"'"},"nodes": {"SS":["'"$REPLICA_ID"'"]}}' \
        --condition-expression 'attribute_not_exists(node_type)' \
        --return-consumed-capacity TOTAL
    PRIMARY_RESULT=$?
    echo "   PRIMARY_RESULT=$PRIMARY_RESULT"

    # make sure everything is good, sleep for a bit, then keep going.
    sleep 30
    echo "Finished installing DTR"
else
    echo " Secondary DTR node"

    # wait for the dynamodb record is available.
    n=0
    until [ $n -ge 30 ]
    do
        echo "Try #$n .."
        REPLICAS=$(aws dynamodb get-item --region $REGION --table-name $DYNAMODB_TABLE --key '{"node_type":{"S": "'"$DTR_DYNAMO_FIELD"'"}}')
        NUM_REPLICAS=$(echo $REPLICAS | jq -r '.Item.nodes.SS | length')
        echo "REPLICAS=$REPLICAS"
        echo "NUM_REPLICAS=$NUM_REPLICAS"
        # if REPLICAS or NUM_REPLICAS is empty or NUM_REPLICAS = 0, it isn't ready sleep
        # and try again.
        if [ -z "$REPLICAS" ] || [ -z "$NUM_REPLICAS" ] || [ $NUM_REPLICAS -eq 0 ]; then
            echo "DTR replicas Not ready yet, sleep for 60 seconds. try #$n"
            sleep 60
            n=$[$n+1]
        else
            echo "DTR replica is ready"
            break
        fi
        if [[ $n -eq 30 ]]; then
            echo "Waiting for DTR replicas timeout! waited to long. start again from the top."
            exit 1
        fi
    done

    sleep 30
    # once available.
    # get record, and then join, add replica ID to dynamodb
    EXISTING_REPLICA_ID=$(echo $REPLICAS | jq -r '.Item.nodes.SS[0]')
    docker run --label com.docker.editions.system --rm "$DTR_IMAGE" join --replica-https-port "$DTR_HTTPS_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$LOCAL_HOSTNAME" --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --existing-replica-id $EXISTING_REPLICA_ID

    JOIN_RESULT=$?
    echo "   JOIN_RESULT=$JOIN_RESULT"
    if [ $JOIN_RESULT -ne 0 ]; then
        echo "We failed for a reason, lets retry again from the top after a brief delay."
        # sleep for a bit first so we give some time for it to recover from the error.
        sleep 30
        exit $JOIN_RESULT
    fi

    # add replica ID so we have a record of it.
    REPLICA_ID=$(docker ps --format '{{.Names}}' -f name=dtr-registry | tail -c 13)
    echo "REPLICA_ID=$REPLICA_ID "
    echo "Not a DTR leader, add secondary manager to dynamodb"
    aws dynamodb update-item \
        --table-name $DYNAMODB_TABLE \
        --region $REGION \
        --key '{"node_type":{"S":  "'"$DTR_DYNAMO_FIELD"'"}}' \
        --update-expression 'ADD nodes :n' \
        --expression-attribute-values '{":n": {"SS":["'"$REPLICA_ID"'"]}}' \
        --return-consumed-capacity TOTAL

fi


echo "Notify AWS that this dtr node is ready"
cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION

#!/bin/bash
echo "#================"
echo "Start DDC setup"

# default to yes if INSTALL DDC is empty.
INSTALL_DDC=${INSTALL_DDC:-"yes"}

PRODUCTION_UCP_ORG='docker'
UCP_ORG=${UCP_ORG:-"docker"}
UCP_TAG=${UCP_TAG-"2.1.5"}
DTR_TAG=${DTR_TAG-"2.2.7"}
DTR_ORG=${DTR_ORG:-"docker"}
UCP_IMAGE=${UCP_ORG}/ucp:${UCP_TAG}
DTR_IMAGE=${DTR_ORG}/dtr:${DTR_TAG}
UCP_HTTPS_PORT=12390
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
echo "UCP_IMAGE=$UCP_IMAGE"
echo "DTR_IMAGE=$DTR_IMAGE"
echo "UCP_HTTPS_PORT=$UCP_HTTPS_PORT"
echo "DTR_HTTPS_PORT=$DTR_HTTPS_PORT"
echo "DTR_HTTP_PORT=$DTR_HTTP_PORT"
echo "MANAGER_COUNT=$MANAGER_COUNT"
echo "#================"

# we don't want to install, exit now.
if [[ "$INSTALL_DDC" != "yes" ]] ; then
  exit 0
fi

if [[ "${UCP_ORG}" == "dockerorcadev" ]] ; then
  IMAGE_LIST_ARGS="--image-version=dev:"
fi

# Login if credentials were provided
if [[ "$REGISTRY_PASSWORD" != "" ]] ; then
  docker login -u "${REGISTRY_USERNAME}" -p "${REGISTRY_PASSWORD}"
fi

if [ "$NODE_TYPE" == "worker" ] ; then
   echo "Let AWS know this worker node is ready."
   docker pull ${UCP_ORG}/ucp-agent:${UCP_TAG}
   cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION
   # nothing else left to do for workers, so exit.
   exit 0
fi

images=$(docker run --rm ${UCP_IMAGE} images --list $IMAGE_LIST_ARGS)
for im in $images; do
  docker pull $im &
done
wait
echo "==> UCP download complete <=="

# Checking if UCP is up and running
checkUCP(){
  echo "Checking to see if UCP is up and healthy"
  n=0
  until [ $n -gt 20 ];
  do
    echo "Checking managers. Try # $n .."
    MANAGERS=$(docker node inspect $(docker node ls --filter role=manager -q) | jq -r '.[] | select(.ManagerStatus.Reachability == "reachable") | .ManagerStatus.Addr | split(":")[0]')
    # Find first node that's not myself
    echo "List of available Managers = $MANAGERS"
    ALLGOOD='yes'
    for I in $MANAGERS; do
      echo "Checking $I to see if UCP is up"
      # Checking if UCP is up and running
      if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$I:$UCP_HTTPS_PORT/_ping) -ne 200 ]] ; then
        echo "   - UCP on $I is NOT healty"
        ALLGOOD='no'
      else
        echo "   + UCP on $I is healthy!"
      fi
    done

    if [[ "$ALLGOOD" == "yes" ]] ; then
      echo "UCP is all healty, good to move on!"
      break
    else
      echo "Not all healthy, rest and try again.."
      if [[ $n -eq 20 ]] ; then
        echo "UCP failed status check after $n tries. Aborting..."
        exit 1
      fi
      sleep 60
      let n+=1
    fi

  done
}

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

IS_UCP_RUNNING=false
checkDDCAgent() {
  echo "Is DDC already running on swarm?"
  # if UCP is running, then this service grep will return a value of 0
  # if not, it will return 1.
  docker service ls | grep /ucp-

  IS_UCP_RUNNING_RESULT=$?  # 0 = installed, 1 = not installed
  echo " IS_UCP_RUNNING_RESULT=$IS_UCP_RUNNING_RESULT"
  if [ $IS_UCP_RUNNING_RESULT -eq 0 ]; then
  echo " UCP Agent is already installed"
  IS_UCP_RUNNING=true
  fi
}

cleanDTR() {
  # It's possible to run into a DTR error, such as: https://github.com/docker/orca/issues/4349 
  # In which case you should
  # 1. Clean up DTR from the node
  # 2. Wait 30s
  # 3. Try to re-install
  echo "Cleaning DTR replica $REPLICA_ID"
  docker run --label com.docker.editions.system  --rm "$DTR_IMAGE" destroy --ucp-url https://$UCP_ELB_HOSTNAME --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --replica-id $REPLICA_ID
  echo "Wait 30s before container restart"
  sleep 30
  exit 1
}

removeDTR() {
  echo "Removing DTR and replica $UNHEALTHY_REPLICA_ID"
  docker run --label com.docker.editions.system --rm "$DTR_IMAGE" remove --ucp-url https://$UCP_ELB_HOSTNAME --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --existing-replica-id $EXISTING_REPLICA_ID --replica-id $UNHEALTHY_REPLICA_ID
}

echo "Wait until we have enough managers up and running."
NUM_MANAGERS=$(docker node inspect $(docker node ls --filter role=manager -q) | jq -r '.[] | select(.ManagerStatus.Reachability == "reachable") | .ManagerStatus.Addr | split(":")[0]' | wc -w)
echo "Current number of Managers = $NUM_MANAGERS"

while [ $NUM_MANAGERS -lt $MANAGER_COUNT ];
do
  echo "Not enough managers yet. We only have $NUM_MANAGERS and we need $MANAGER_COUNT to continue."
  echo "sleep for a bit, and try again when we wake up."
  sleep 30
  NUM_MANAGERS=$(docker node inspect $(docker node ls --filter role=manager -q) | jq -r '.[] | select(.ManagerStatus.Reachability == "reachable") | .ManagerStatus.Addr | split(":")[0]' | wc -w)
  # if we never get to the number of managers, the stack will timeout, so we don't have to worry
  # about being stuck in the loop forever.
done

echo "We have enough managers we can continue now."

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
  echo "We are the swarm leader"
  echo "Installing DDC..."

  # Loading the License
  echo "Loading DDC License"
  if [[ ${LICENSE:0:1} == "{"  ]];
    then echo "Using JSON Direct Input"
    echo $LICENSE >> /tmp/docker/docker_subscription.lic
    IS_VALID_LICENSE=1
  elif [[  ${LICENSE:0:4} == "http" ]];
    then echo "Using URL to download license file"
    curl -s $LICENSE >> /tmp/docker/docker_subscription.lic
    IS_VALID_LICENSE=1
  else echo "License input must be a valid URL or JSON license key. Please upload license in UI after installation."
    IS_VALID_LICENSE=0
  fi

  # Check if UCP is already installed, if it is exit. If not, install it.
  if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$UCP_ELB_HOSTNAME/_ping) -ne 200 ]]; then
    # Installing UCP
    echo "Run the UCP install script"
    if [[ ${IS_VALID_LICENSE} -eq 1 ]]; then
      docker run --label com.docker.editions.system  --rm --name ucp -v /tmp/docker/docker_subscription.lic:/config/docker_subscription.lic -v /var/run/docker.sock:/var/run/docker.sock "$UCP_IMAGE" install ${UCP_IMAGE_DEV_FLAG} --controller-port "$UCP_HTTPS_PORT" --san "$UCP_ELB_HOSTNAME" --external-service-lb "$APP_ELB_HOSTNAME" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
      if [ $? -ne 0 ]; then
        echo "Error installing UCP"
        exit 1
      fi
      echo "Finished installing UCP with license"
    else
      docker run --label com.docker.editions.system  --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock "$UCP_IMAGE" install --san "$UCP_ELB_HOSTNAME" --controller-port "$UCP_HTTPS_PORT" --external-service-lb "$APP_ELB_HOSTNAME" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
      if [ $? -ne 0 ]; then
        echo "Error installing UCP"
        exit 1
      fi
      echo "Finished installing UCP without license. Please upload your license in UCP and DTR UI. "
    fi
  else
    echo "UCP is already installed, continue to DTR"
  fi

  # make sure UCP is ready before we continue
  checkUCP

  # wait for a full two minutes before attempting to install DTR
  echo "Waiting for things to settle down before installing DTR"
  sleep 120

  # Checking if DTR is already running. If it is , exit, if it's not install it.
  if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$MYIP:$DTR_HTTPS_PORT/health) -ne 200 ]]; then

    # For upgrades, ensure that DTR isn't already installed
    REPLICAS=$(aws dynamodb get-item --region $REGION --table-name $DYNAMODB_TABLE --key '{"node_type":{"S": "'"$DTR_DYNAMO_FIELD"'"}}')
    NUM_REPLICAS=$(echo $REPLICAS | jq -r '.Item.nodes.SS | length')
    # if we get a result, we know DTR is already running on this cluster

    if [[ $NUM_REPLICAS -eq 0 ]] ; then
      echo "Generate our DTR replica ID"
      # create a unique replica id, given the IP address of this host.
      REPLICA_ID=$(echo $MYIP | sed "s/\./0/g" | awk '{print "0000"$1}' | tail -c 13)
      echo "REPLICA_ID=$REPLICA_ID "
      DTR_LEADER_INSTALL="yes"
      echo "Installing First DTR Replica..."
      sleep 30
      echo "Install DTR"
      date
      docker run --label com.docker.editions.system  --rm "$DTR_IMAGE" install --replica-https-port "$DTR_HTTPS_PORT" --replica-http-port "$DTR_HTTP_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$NODE_NAME" --dtr-external-url $DTR_ELB_HOSTNAME:443 --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --replica-id $REPLICA_ID
      INSTALL_RESULT=$?
      echo "   INSTALL_RESULT=$JOIN_RESULT"
      if [ $INSTALL_RESULT -ne 0 ];
      then
        echo "We failed for a reason, lets retry again from the top after a brief delay."
        cleanDTR $REPLICA_ID
        exit $INSTALL_RESULT
      fi
      echo "After running install via Docker"
      date
      # make sure everything is good, sleep for a bit, then keep going.
      sleep 30
      echo "Finished installing DTR"
    else
      # lets make sure everything is good to go, before we get started.
      # If we start too soon, it could choke. random sleep so we don't step
      # on other nodes that are also joining
      sleep $[ ( $RANDOM % 30 )  + 11 ]
      echo "DTR already installed, need to join instead of install"
      DTR_LEADER_INSTALL="no"
      EXISTING_REPLICA_ID=$(echo $REPLICAS | jq -r '.Item.nodes.SS[0]')
      echo "Join to replicaId = $EXISTING_REPLICA_ID"
      CURRENT_DTR=$(curl --silent http://$MYIP:9024/info/dtr/)
      if [[ -n "$CURRENT_DTR" && "$CURRENT_DTR" != "404: Page not found" ]]; then
        DTR_IMAGE=${DTR_ORG}/dtr:${CURRENT_DTR}
      fi
      docker run --label com.docker.editions.system  --rm "$DTR_IMAGE" join --replica-https-port "$DTR_HTTPS_PORT" --replica-http-port "$DTR_HTTP_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$LOCAL_HOSTNAME" --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --existing-replica-id $EXISTING_REPLICA_ID
    fi
  else
    echo "DTR already running"
    echo "Notify AWS that this manager node is ready"
    cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION
    echo "Finished.."
    exit 0
  fi

  checkDTR

  if [[ "$DTR_LEADER_INSTALL" == "yes" ]] ; then
    echo "Is a DTR leader. Add replicaID to dynamodb"
    aws dynamodb put-item \
      --table-name $DYNAMODB_TABLE \
      --region $REGION \
      --item '{"node_type":{"S":  "'"$DTR_DYNAMO_FIELD"'"},"nodes": {"SS":["'"$REPLICA_ID"'"]}}' \
      --condition-expression 'attribute_not_exists(node_type)' \
      --return-consumed-capacity TOTAL
    PRIMARY_RESULT=$?
    # If this fails, we simply didn't update our DynamoDB with the leader - Can make other deployment fail
    echo "   PRIMARY_RESULT=$PRIMARY_RESULT"

    # Configuring DTR with S3
    echo "Configuring DTR with S3 Storage Backend..."
    SET_BUCKET=$(jq -n --arg REGION "$REGION" --arg S3_BUCKET_NAME "$S3_BUCKET_NAME" '
{
  "storage": {
  "delete": {
    "enabled": true
  },
  "maintenance": {
    "readonly": {
    "enabled": false
    }
  },
  "s3": {
    "rootdirectory": "",
    "region": $REGION,
    "regionendpoint": "",
    "bucket": $S3_BUCKET_NAME,
    "secure": true
  }
  }
}')
    S3_STATUS=$(curl -H 'content-type: application/json' --silent --output /dev/null --write-out '%{http_code}' -k -u $UCP_ADMIN_USER:$UCP_ADMIN_PASSWORD -X PUT "https://$DTR_ELB_HOSTNAME/api/v0/admin/settings/registry/simple" -d "$SET_BUCKET")
    if [[ $S3_STATUS -lt 300 ]];
      then echo "Successfully Configured DTR storage backend with S3"
    else
      echo "Failed to configure DTR storage backend with S3"
      # Additional Debugging Info:
      curl -v -H 'content-type: application/json' --write-out '%{http_code}' -k -u $UCP_ADMIN_USER:$UCP_ADMIN_PASSWORD -X PUT "https://$DTR_ELB_HOSTNAME/api/v0/admin/settings/registry/simple" -d "$SET_BUCKET"
    fi
  else
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
  echo "Notify AWS that this manager node is ready"
  cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION

else
  echo "Not the Swarm leader. "
  echo "Notify AWS that this manager node is ready - we will get UCP via the agent, need to install DTR after"
  cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION

  # make sure UCP is ready.
  checkUCP

  echo "UCP is ready, lets install DTR now."
  # DTR stuff here.
  # check to see if dtr is already installed. if not continue
  # Checking if DTR is already running. If it is , exit, if it's not install it.
  if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$MYIP:$DTR_HTTPS_PORT/health) -ne 200 ]]; then
    echo "install DTR"

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
    CURRENT_DTR=$(curl --silent http://$MYIP:9024/info/dtr/)
    if [[ -n "$CURRENT_DTR" && "$CURRENT_DTR" != "404: Page not found" ]]; then
      DTR_IMAGE=${DTR_ORG}/dtr:${CURRENT_DTR}
    fi
    docker run --label com.docker.editions.system --rm "$DTR_IMAGE" join --replica-https-port "$DTR_HTTPS_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$LOCAL_HOSTNAME" --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --existing-replica-id $EXISTING_REPLICA_ID

    JOIN_RESULT=$?
    echo "   JOIN_RESULT=$JOIN_RESULT"
    if [ $JOIN_RESULT -ne 0 ]; then
      echo "We failed for a reason, lets retry again from the top after a brief delay."
      # sleep for a bit first so we give some time for it to recover from the error.
      sleep 30
      exit $JOIN_RESULT
    fi

    # check to make sure that DTR is ready
    checkDTR

    REPLICA_ID=$(docker ps --format '{{.Names}}' -f name=dtr-registry | tail -c 13)
    echo "REPLICA_ID=$REPLICA_ID "

    echo "DTR replica ID to dynamodb"
    aws dynamodb update-item \
      --table-name $DYNAMODB_TABLE \
      --region $REGION \
      --key '{"node_type":{"S": "'"$DTR_DYNAMO_FIELD"'"}}' \
      --update-expression 'ADD nodes :n' \
      --expression-attribute-values '{":n": {"SS":["'"$REPLICA_ID"'"]}}' \
      --return-consumed-capacity TOTAL
  else
    echo "DTR is already installed.."
  fi

fi

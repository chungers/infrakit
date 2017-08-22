#!/bin/bash
echo "#================"
echo "Start DDC setup"

# default to yes if INSTALL DDC is empty.
INSTALL_DDC=${INSTALL_DDC:-"yes"}

PRODUCTION_UCP_ORG='docker'
UCP_ORG=${UCP_ORG:-"docker"}
UCP_TAG=${UCP_TAG-"2.1.5"}
UCP_IMAGE=${UCP_ORG}/ucp:${UCP_TAG}
DTR_TAG=${DTR_TAG-"2.2.7"}
DTR_ORG=${DTR_ORG:-"docker"}
DTR_IMAGE=${DTR_ORG}/dtr:${DTR_TAG}

UCP_HTTPS_PORT=12390
DTR_HTTPS_PORT=12391
DTR_HTTP_PORT=12392
DTR_SEQ_ID=0
IMAGE_LIST_ARGS=''

echo "APP_ID=$APP_ID"
echo "TENANT_ID=$TENANT_ID"
echo "ACCOUNT_ID=$ACCOUNT_ID"
echo "PATH=$PATH"
echo "ROLE=$ROLE"
echo "REGION=$REGION"
echo "GROUP_NAME=$GROUP_NAME"
echo "UCP_LICENSE=$UCP_LICENSE"
echo "INSTALL_DDC=$INSTALL_DDC"
echo "AZURE_HOSTNAME=$AZURE_HOSTNAME"
echo "APP_ELB_HOSTNAME=$APP_ELB_HOSTNAME"
echo "UCP_ELB_HOSTNAME=$UCP_ELB_HOSTNAME"
echo "DTR_ELB_HOSTNAME=$DTR_ELB_HOSTNAME"
echo "UCP_ADMIN_USER=$UCP_ADMIN_USER"
echo "UCP_IMAGE=$UCP_IMAGE"
echo "DTR_IMAGE=$DTR_IMAGE"
echo "UCP_HTTPS_PORT=$UCP_HTTPS_PORT"
echo "DTR_HTTPS_PORT=$DTR_HTTPS_PORT"
echo "DTR_HTTP_PORT=$DTR_HTTP_PORT"
echo "DTR_STORAGE_ACCOUNT=$DTR_STORAGE_ACCOUNT"
echo "PRIVATE_IP=$PRIVATE_IP"
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

#Download Docker UCP images
images=$(docker run --label com.docker.editions.system --rm ${UCP_IMAGE} images --list $IMAGE_LIST_ARGS)
for im in $images; do
    docker pull $im
done

#Download DTR images
images=$(docker run --rm $DTR_IMAGE images)
for im in $images; do
    docker pull $im
done


if [ "$NODE_TYPE" == "worker" ] ; then
  # nothing left to do for workers, so exit.
  exit 0
fi

#Get vmss node count
mgr_node_count=$(python dtrutils.py get-mgr-nodes)
echo "VMSS Manager Node Count: $mgr_node_count"
wrk_node_count=$(python dtrutils.py get-wrk-nodes)
echo "VMSS Worker Node Count: $wrk_node_count"

echo "Wait until we have enough managers up and running."
num_managers=$(docker node inspect $(docker node ls --filter role=manager -q) | jq -r '.[] | select(.ManagerStatus.Reachability == "reachable") | .ManagerStatus.Addr | split(":")[0]' | wc -w)
echo "Current number of Managers = $num_managers"

while [ $num_managers -lt $mgr_node_count ];
do
  echo "Not enough managers yet. We only have $num_managers and we need $mgr_node_count to continue."
  echo "sleep for a bit, and try again when we wake up."
  sleep 30
  num_managers=$(docker node inspect $(docker node ls --filter role=manager -q) | jq -r '.[] | select(.ManagerStatus.Reachability == "reachable") | .ManagerStatus.Addr | split(":")[0]' | wc -w)

  # if we never get to the number of managers, the stack will timeout, so we don't have to worry
  # about being stuck in the loop forever.
done

echo "We have enough managers we can continue now."

# Check to make sure swarm nodes are up  before proceeding

num_swarm_nodes=$(($mgr_node_count + $wrk_node_count))
echo "Total Number of Nodes:" $num_swarm_nodes
COUNTER=0
while :
do
  active_nodes=$(docker node ls | grep Ready | awk '{print $2}' | wc -l)
  echo "Num of Swarm Nodes up:" $active_nodes
  if [ $active_nodes -lt $num_swarm_nodes ]
  then
    sleep 30 
  else
    break
  fi
  COUNTER=$((COUNTER + 1))
   if [ $COUNTER -gt 60 ]
  then
    echo "Issues with Swarm setup -- Please delete Resource group and redeploy the template"
    exit 1
  fi
done

#Check and remove failed Nodes that are not part of the Swarm setup

echo "check and remove failed nodes not part of swarm "
for node_id in $(docker node ls | grep -E "Unknown|Down" | awk '{print $1}');
do
  echo "node_id: $node_id "
  if [[ $node_id != "" ]] ;
  then
    docker node rm $node_id
  fi
done

echo "Swarm Cluster is up."

# Checking if UCP is up and running
checkUCP(){
  echo "Checking to see if UCP is up"
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
      if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://"$I":$UCP_HTTPS_PORT/_ping) -ne 200 ]] ; then
        echo "   - UCP on $I is NOT healthy"
        ALLGOOD='no'
      else
        echo "   + UCP on $I is healthy!"
      fi
    done

    if [[ "$ALLGOOD" == "yes" ]] ; then
      echo "UCP is all healthy, good to move on!"
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

#Check UCP ELB Health
checkUCP_ELB(){
  COUNT=0
  while :
  do
    if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://"$UCP_ELB_HOSTNAME"/_ping) -ne 200 ]]
    then
      echo "sleep and check UCP ELB again"
      sleep 10
    else
      echo "UCP ELB Healthy"
      break
    fi
    COUNT=$((COUNT + 1))
    if [ $COUNT -gt 5 ]
    then
      echo "UCP ELB is unhealthy"
      break
    fi

  done
}


# Checking if DTR is up
checkDTR(){
  echo "Checking to see if DTR is up and healthy"
  n=0
  until [ $n -gt 20 ];
  do
    if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://"$PRIVATE_IP":$DTR_HTTPS_PORT/health) -eq 200 ]];
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
}

removeDTR() {
    echo "Removing DTR and replica $UNHEALTHY_REPLICA_ID"
    docker run --label com.docker.editions.system --rm "$DTR_IMAGE" remove --ucp-url https://$UCP_ELB_HOSTNAME --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --existing-replica-id $EXISTING_REPLICA_ID --replica-id $UNHEALTHY_REPLICA_ID
    REMOVE_RESULT=$?
    echo "REMOVE_RESULT=$REMOVE_RESULT"
    if [ $REMOVE_RESULT -eq 0 ]; then
      python dtrutils.py delete-id $UNHEALTHY_REPLICA_ID
      echo "deleted Replica ID (leader section):$UNHEALTHY_REPLICA_ID"
    fi
}

# Check and set DTR flag to run DTR sequentially 
checkDTR_flag(){
  #DTR replicas need to be run sequentially
  # Add a random delay so they don't step on each other during DTR install.
  echo "Sleep for a short time (1-30 seconds). To prevent scripts from stepping on each other"
  sleep $[ ( $RANDOM % 30 )  + 11 ]
  # echo "Finished sleep, lets get going."
  echo "$(date) $line"

  seq_id=$(python dtrutils.py get-id)
  echo "seq_id: $seq_id"
  if [ $seq_id -ne 0 ]; then
    echo "sleeping for seq id "
    n=0
    until [ $n -ge 40 ]
    do
      echo "sleeping for seq id "
      sleep 30
      seq_id=$(python dtrutils.py get-id)
      echo "seq_id: $seq_id"
      if [ $seq_id -ne 0 ] ; then
        sleep 30
        n=$[$n+1]
      else
        echo "update id for next DTR replica join"
        break
      fi
    done
  fi

  DTR_SEQ_ID=1
  python dtrutils.py add-id $DTR_SEQ_ID
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

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]] && [[ "$IS_UCP_RUNNING" == "false" ]]  ; then
  echo "We are the swarm leader"
  echo "Setup DDC"
  
  # Loading DDC  License
  echo "Loading DDC License"
  if [[ $UCP_LICENSE != "" ]];
  then
    LIC_FILE=/tmp/docker/docker_subscription.lic
    echo -n  "$UCP_LICENSE" | base64 -d >> $LIC_FILE
    LICENSE=`cat $LIC_FILE`
    jq -e '.|{key_id}' $LIC_FILE >> /dev/null
    if [[ $? -eq 0 ]]
    then
      echo "valid license "
      IS_VALID_LICENSE=1
    else 
      echo "License input must be a valid JSON license key. Please upload license in UI after installation." 
      IS_VALID_LICENSE=0
    fi 
  else
    echo "Unable to read license file. Please upload license in UI after installation."
  fi
  
  # Check if UCP is already installed, if it is exit. If not, install it.
  if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://"$UCP_ELB_HOSTNAME"/_ping) -ne 200 ]];
  then 
  # Installing UCP 
    echo "Run the UCP install script" 
    if [[ ${IS_VALID_LICENSE} -eq 1 ]];
    then 
      docker run --label com.docker.editions.system --rm --name ucp -v /tmp/docker/docker_subscription.lic:/config/docker_subscription.lic -v /var/run/docker.sock:/var/run/docker.sock ${UCP_IMAGE} install --controller-port $UCP_HTTPS_PORT --san "$UCP_ELB_HOSTNAME" --external-service-lb "$APP_ELB_HOSTNAME" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
      echo "Finished installing UCP with license"
    else
      docker run --label com.docker.editions.system --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock ${UCP_IMAGE} install --controller-port $UCP_HTTPS_PORT --san "$UCP_ELB_HOSTNAME" --external-service-lb "$APP_ELB_HOSTNAME" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
      echo "Finished installing UCP without license. Please upload your license in UCP and DTR UI. "
    fi
  else
    echo "UCP is already installed, continue to DTR"
  fi

  # Checking if UCP is up and running
  checkUCP 

  #Check if UCP ELB is healthy
  checkUCP_ELB

  # Checking if DTR is already running. If it is , exit, if it's not install it. 
  if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://"$PRIVATE_IP":$DTR_HTTPS_PORT/health) -ne 200 ]]; 
  then
    # For upgrades, ensure that DTR isn't already installed
    REPLICAS=$(python dtrutils.py get-ids)
    echo "REPLICAS: $REPLICAS"
    NUM_REPLICAS=0
    REPLICA_ID=0
    UNHEALTHY_REPLICA_ID=0
    for id in $REPLICAS
    do
      echo "Replica ID: $id"
      NUM_REPLICAS=$((NUM_REPLICAS + 1))
      node_name=$(python dtrutils.py get-nodename $id)
      echo "node name: $node_name"
      node_id=$(docker node ls --filter name=$node_name -q)
      echo "node id: $node_id"
      if [[ $node_id == "" ]]; then
        UNHEALTHY_REPLICA_ID=$id
      else	
        EXISTING_REPLICA_ID=$id
      fi
    done
    echo "Number of REPLICAS: $NUM_REPLICAS"

    # if we get a result, we know DTR is already running on this cluster

    if [[ $NUM_REPLICAS -eq 0 ]] ; then
      echo "Generate our DTR replica ID"
      # create a unique replica id, given the IP address of this host.
      REPLICA_ID=$(echo $PRIVATE_IP | sed "s/\./0/g" | awk '{print "0000"$1}' | tail -c 13)
      echo "REPLICA_ID=$REPLICA_ID "
      DTR_LEADER_INSTALL="yes"
      # create Azure Table storage
      python dtrutils.py create-table
      echo "Installing First DTR Replica..."
      sleep 20
      echo "Install DTR"
      date
      echo "AZURE_HOSTNAME=$AZURE_HOSTNAME"
      docker run --label com.docker.editions.system --rm "$DTR_IMAGE" install --replica-https-port "$DTR_HTTPS_PORT" --replica-http-port "$DTR_HTTP_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$AZURE_HOSTNAME" --dtr-external-url $DTR_ELB_HOSTNAME --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --replica-id $REPLICA_ID
      INSTALL_RESULT=$?
      echo " INSTALL_RESULT=$INSTALL_RESULT"
      if [ $INSTALL_RESULT -ne 0 ]; then
        echo "We failed for a reason, lets retry again from the top after a brief delay."
        cleanDTR $REPLICA_ID
        exit $INSTALL_RESULT
      fi
      echo "After running install via Docker"
      echo "set flag to serialize DTR replica join "
      python dtrutils.py add-id $DTR_SEQ_ID
      # insert Replica ID in Azure Table
      python dtrutils.py insert-id $REPLICA_ID $AZURE_HOSTNAME "INITIAL INSTALL"
      date
      # make sure everything is good, sleep for a bit, then keep going.
      sleep 30
      echo "Finished installing DTR"
    else
      # lets make sure everything is good to go, before we get started.
      # If we start too soon, it could choke. random sleep so we don't step
      # on other nodes that are also joining
      # cleanup unhealthy DTR
      echo "UNHEALTHY_REPLICA_ID: $UNHEALTHY_REPLICA_ID"
      if [[ $UNHEALTHY_REPLICA_ID -ne 0 ]]; then
        removeDTR $UNHEALTHY_REPLICA_ID
        
      fi
      sleep $[ ( $RANDOM % 30 )  + 11 ]
      echo "DTR already installed, need to join instead of install"
      DTR_LEADER_INSTALL="no"
      checkDTR_flag
      echo "$AZURE_HOSTNAME node processing DTR"	
      echo "Join to replicaId = $EXISTING_REPLICA_ID"
      docker run --label com.docker.editions.system --rm "$DTR_IMAGE" join --replica-https-port "$DTR_HTTPS_PORT" --replica-http-port "$DTR_HTTP_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$AZURE_HOSTNAME" --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --existing-replica-id $EXISTING_REPLICA_ID
      JOIN_RESULT=$?
      echo "   JOIN_RESULT=$JOIN_RESULT"
      if [ $JOIN_RESULT -ne 0 ]; then
        echo "We failed for a reason, lets retry again from the top after a brief delay."
        # sleep for a bit first so we give some time for it to recover from the error.
        sleep 20
        DTR_SEQ_ID=0
        python dtrutils.py add-id $DTR_SEQ_ID
        echo "Join failed, retrying again"
        exit $JOIN_RESULT
      fi
      sleep 10
      DTR_SEQ_ID=0
      python dtrutils.py add-id $DTR_SEQ_ID

    fi
  else
    echo "DTR already running"
    echo "Finished.."
    exit 0
  fi

  # Checking if DTR is up
  checkDTR

  if [[ "$DTR_LEADER_INSTALL" == "yes" ]] ; then

    #Configure Azure Storage
    DTR_STORAGE_KEY=$(python dtrutils.py get-key)
    echo "DTR STORAGE KEY: $DTR_STORAGE_KEY"

    if [[ $(curl --silent --output /dev/null --write-out '%{http_code}' -k -u $UCP_ADMIN_USER:$UCP_ADMIN_PASSWORD -X PUT "https://$DTR_ELB_HOSTNAME/api/v0/admin/settings/registry/simple" -d "{\"storage\":{\"delete\":{\"enabled\":true},\"maintenance\":{\"readonly\":{\"enabled\":false}},\"azure\":{\"accountname\":\"$DTR_STORAGE_ACCOUNT\", \"accountkey\":\"$DTR_STORAGE_KEY\", \"container\":\"dtrcontainer\"}}}") -lt 300 ]];
    then
      echo " Successfully Configured DTR storage backend with Azure Blob "
    else
      echo " Failed to configure DTR storage backend. Please configure BLOB storage from DTR UI (under settings) "
      #Additional Debugging Info
      curl -v --write-out '%{http_code}' -k -u $UCP_ADMIN_USER:$UCP_ADMIN_PASSWORD -X PUT "https://$DTR_ELB_HOSTNAME/api/v0/admin/settings/registry/simple" -d "{\"storage\":{\"delete\":{\"enabled\":true},\"maintenance\":{\"readonly\":{\"enabled\":false}},\"azure\":{\"accountname\":\"$DTR_STORAGE_ACCOUNT\", \"accountkey\":\"$DTR_STORAGE_KEY\", \"container\":\"dtrcontainer\"}}}"
    fi
  else
    REPLICA_ID=$(docker ps --format '{{.Names}}' -f name=dtr-registry | tail -c 13)
    echo "REPLICA_ID=$REPLICA_ID "
    echo "Not a DTR leader, add secondary manager to Azure Table"
    python dtrutils.py insert-id $REPLICA_ID $AZURE_HOSTNAME "INITIAL INSTALL"
  fi

else
  echo "Not the swarm leader. "

  #make sure UCP is ready.
  checkUCP

  echo "UCP is ready, lets install DTR now."
  # DTR stuff here.
  # check to see if dtr is already installed. if not continue
  # Checking if DTR is already running. If it is , exit, if it's not install it.
  if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$PRIVATE_IP:$DTR_HTTPS_PORT/health) -ne 200 ]]; then
    echo "install DTR"

    # Avoid installing DTR during upgrade process, causes issues
    COUNTER=0
    while :
    do
      desc=$(python dtrutils.py get-desc)
      echo "Description: $desc"
      if [[ $desc == "upgrade in progress" ]]
      then
        echo "upgrade in process -- try again later"
        sleep 10
      else
        echo "Go ahead with DTR install/join"
        break
      fi
      COUNTER=$((COUNTER + 1))
      if [ $COUNTER -gt 500 ]
      then
        echo "upgrade process is taking too long"
        exit 1
      fi
    done

    # wait till Azure Table record is available.
    n=0
    until [ $n -ge 30 ]
    do
      echo "Try #$n .."
      REPLICAS=$(python dtrutils.py get-ids)
      echo "REPLICAS=$REPLICAS"
      NUM_REPLICAS=0
      for id in $REPLICAS
      do
        echo "Replica ID: $id"
        NUM_REPLICAS=$((NUM_REPLICAS + 1))
      done
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
    # get record, and then join, add replica ID to Azure table
    EXISTING_REPLICA_ID=${REPLICAS[0]}
    echo "Existing REPLICA ID: $EXISTING_REPLICA_ID"

    # check flag before DTR join is run
    checkDTR_flag

    echo "$AZURE_HOSTNAME node processing DTR"

    docker run --label com.docker.editions.system --rm "$DTR_IMAGE" join --replica-https-port "$DTR_HTTPS_PORT" --replica-http-port "$DTR_HTTP_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$AZURE_HOSTNAME" --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --existing-replica-id $EXISTING_REPLICA_ID

    JOIN_RESULT=$?
    echo "   JOIN_RESULT=$JOIN_RESULT"
    if [ $JOIN_RESULT -ne 0 ]; then
      echo "We failed for a reason, lets retry again from the top after a brief delay."
      # sleep for a bit first so we give some time for it to recover from the error.
      sleep 30
      DTR_SEQ_ID=0
      python dtrutils.py add-id $DTR_SEQ_ID
      echo "Join failed, retrying again"
      exit $JOIN_RESULT
    fi
    sleep 10
    DTR_SEQ_ID=0
    python dtrutils.py add-id $DTR_SEQ_ID
    echo "SEQ id updated"

    # check to make sure that DTR is ready
    checkDTR

    REPLICA_ID=$(docker ps --format '{{.Names}}' -f name=dtr-registry | tail -c 13)
    echo "REPLICA_ID=$REPLICA_ID "

    echo "DTR replica ID to Azure Table"
    python dtrutils.py insert-id $REPLICA_ID $AZURE_HOSTNAME "INITIAL INSTALL"
  else
    echo "DTR is already installed.."
  fi
fi

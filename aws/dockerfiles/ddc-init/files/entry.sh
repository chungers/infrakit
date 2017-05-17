#!/bin/bash
echo "#================"
echo "Start DDC setup"

# default to yes if INSTALL DDC is empty.
INSTALL_DDC=${INSTALL_DDC:-"yes"}

PRODUCTION_UCP_ORG='docker'
UCP_ORG=${UCP_ORG:-"docker"}
UCP_TAG=${UCP_TAG-"2.1.1"}
UCP_IMAGE=${UCP_ORG}/ucp:${UCP_TAG}
UCP_HTTPS_PORT=12390
IMAGE_LIST_ARGS=''
MYIP=$(wget -qO- http://169.254.169.254/latest/meta-data/local-ipv4)
LOCAL_HOSTNAME=$(wget -qO- http://169.254.169.254/latest/meta-data/local-hostname)

# Normalize the ELB hostnames to be all lowercase.
APP_ELB_HOSTNAME=$(tr '[:upper:]' '[:lower:]' <<< "$APP_ELB_HOSTNAME")
UCP_ELB_HOSTNAME=$(tr '[:upper:]' '[:lower:]' <<< "$UCP_ELB_HOSTNAME")

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
echo "NODE_NAME=$NODE_NAME"
echo "UCP_ADMIN_USER=$UCP_ADMIN_USER"
echo "UCP_IMAGE=$UCP_IMAGE"
echo "UCP_HTTPS_PORT=$UCP_HTTPS_PORT"
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

images=$(docker run --label com.docker.editions.system  --rm "${UCP_ORG}/ucp:${UCP_TAG}" images --list $IMAGE_LIST_ARGS)
for im in $images; do
    docker pull $im
done

# ^^ If this is an upgrade, do we still need to pull down these images? Should they instead be the version that installed in the swarm?

if [ "$NODE_TYPE" == "worker" ] ; then
     echo "Let AWS know this worker node is ready."
     cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION
     # nothing else left to do for workers, so exit.
     exit 0
fi


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

addDTRToken() {
    echo "Adding DTR-agent token"
    AGENT_USER="dtrapi"
    head -c 24 /dev/urandom | base64 | tee /tmp/dtr.txt | docker secret create dtr_api_token -
    AGENT_PASSWORD=$(cat /tmp/dtr.txt)
    STATUS=$(curl --insecure --silent -u $UCP_ADMIN_USER:$UCP_ADMIN_PASSWORD --output /dev/null --write-out '%{http_code}' -X POST -H 'Content-Type: application/json' -d '{"name": "$AGENT_USER", "password": "$AGENT_PASSWORD", "isAdmin": true, "isActive": true}' "https://$UCP_ELB_HOSTNAME:$UCP_HTTPS_PORT/accounts/")
    if [ "$STATUS" -ne "200" ]; then
        echo "ERROR: Adding user failed"
    fi
    rm -f /tmp/dtr.txt
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

echo "Is DDC already running on swarm?"
# if UCP is running, then this service grep will return a value of 0
# if not, it will return 1.
docker service ls | grep /ucp-

IS_UCP_RUNNING_RESULT=$?  # 0 = installed, 1 = not installed
echo " IS_UCP_RUNNING_RESULT=$IS_UCP_RUNNING_RESULT"
if [ $IS_UCP_RUNNING_RESULT -ne 0 ]; then
    echo "  UCP is not already installed"
    IS_UCP_RUNNING='false'
else
    echo "  UCP is already installed"
    IS_UCP_RUNNING='true'
fi

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]] && [[ "$IS_UCP_RUNNING" == "false" ]]  ; then
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
            echo "Finished installing UCP with license"
        else
            docker run --label com.docker.editions.system  --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock "$UCP_IMAGE" install --san "$UCP_ELB_HOSTNAME" --controller-port "$UCP_HTTPS_PORT" --external-service-lb "$APP_ELB_HOSTNAME" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
            echo "Finished installing UCP without license. Please upload your license in UCP UI. "
        fi
        addDTRToken
    else
        echo "UCP is already installed"
    fi
else
    echo "Not the Swarm leader, or UCP already installed."
fi

# make sure UCP is ready before we continue
checkUCP

echo "Notify AWS that this manager node is ready"
cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION

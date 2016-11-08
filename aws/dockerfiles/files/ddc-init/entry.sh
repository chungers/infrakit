#!/bin/bash
echo "#================"
echo "Start DDC setup"

# default to yes if INSTALL DDC is empty.
INSTALL_DDC=${INSTALL_DDC:-"yes"}

PRODUCTION_HUB_NAMESPACE='docker'
HUB_NAMESPACE=${HUB_NAMESPACE:-"docker"}
UCP_HUB_TAG=${UCP_HUB_TAG-"2.0.0-beta4"}
DTR_HUB_TAG=${DTR_HUB_TAG-"2.1.0-beta4"}
UCP_IMAGE=${HUB_NAMESPACE}/ucp:${UCP_HUB_TAG}
DTR_IMAGE=${HUB_NAMESPACE}/dtr:${DTR_HUB_TAG}
DTR_PORT=8443
IMAGE_LIST_ARGS=''


echo "PATH=$PATH"
echo "STACK_NAME=$STACK_NAME"
echo "REGION=$REGION"
echo "S3_BUCKET_NAME=$S3_BUCKET_NAME"
echo "LICENSE=$LICENSE"
echo "INSTALL_DDC=$INSTALL_DDC"
echo "UCP_ELB_HOSTNAME=$UCP_ELB_HOSTNAME"
echo "DTR_ELB_HOSTNAME=$DTR_ELB_HOSTNAME"
echo "NODE_NAME=$NODE_NAME"
echo "UCP_ADMIN_USER=$UCP_ADMIN_USER"
echo "UCP_IMAGE=$UCP_IMAGE"
echo "DTR_IMAGE=$DTR_IMAGE"
echo "DTR_PORT=$DTR_PORT"
echo "MANAGER_COUNT=$MANAGER_COUNT"
echo "#================"

# we don't want to install, exit now.
if [[ "$INSTALL_DDC" != "yes" ]] ; then
    exit 0
fi

# Loading Beta Images without login
# TODO : Remove this step when DTR+UCP go GA
curl -o docker-datacenter.tar.gz https://packages.docker.com/caas/ucp-2.0.0-beta4_dtr-2.1.0-beta4.tar.gz  && docker load -i docker-datacenter.tar.gz && rm docker-datacenter.tar.gz

# TODO: Add this section back when UCP goes GA
#images=$(docker run --rm "${HUB_NAMESPACE}/ucp:${UCP_HUB_TAG}" images --list $IMAGE_LIST_ARGS )
#for im in $images; do
#    docker pull $im
#done

if [ "$NODE_TYPE" == "worker" ] ; then
     echo "Let AWS know this worker node is ready."
     cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION
     # nothing else left to do for workers, so exit.
     exit 0
fi

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
    if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$UCP_ELB_HOSTNAME/_ping) -ne 200 ]];
        # Installing UCP
        then echo "Run the UCP install script"
        if [[ ${IS_VALID_LICENSE} -eq 1 ]];
            then docker run --rm --name ucp -v /tmp/docker/docker_subscription.lic:/config/docker_subscription.lic -v /var/run/docker.sock:/var/run/docker.sock "$UCP_IMAGE" install --san "$UCP_ELB_HOSTNAME" --external-service-lb "$UCP_ELB_HOSTNAME" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
            echo "Finished installing UCP with license"
        else
            docker run --rm --name ucp -v /var/run/docker.sock:/var/run/docker.sock "$UCP_IMAGE" install --san "$UCP_ELB_HOSTNAME" --external-service-lb "$UCP_ELB_HOSTNAME" --admin-username "$UCP_ADMIN_USER" --admin-password "$UCP_ADMIN_PASSWORD" $IMAGE_LIST_ARGS
            echo "Finished installing UCP without license. Please upload your license in UCP and DTR UI. "
        fi
    else
        exit 0
    fi

    # Checking if UCP is up and running
    echo "Checking to see if UCP is up and healthy"
    checkUCP(){
        MANAGERS=$(docker node inspect $(docker node ls --filter role=manager -q) | jq -r '.[] | select(.ManagerStatus.Reachability == "reachable") | .ManagerStatus.Addr | split(":")[0]')
        # Find first node that's not myself
        echo "List of available Managers = $MANAGERS"
        n=0
        until [ $n -gt 20 ];
        do
            echo "Checking managers. Try # $n .."
            ALLGOOD='yes'
            for I in $MANAGERS; do
                echo "Checking $I to see if UCP is up"
                # Checking if UCP is up and running
                if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$I/_ping) -ne 200 ]] ; then
                    echo "UCP on $I is NOT healty"
                    ALLGOOD='no'
                else
                    echo "UCP on $I is healthy!"
                fi
            done

            if [[ "$ALLGOOD" == "yes" ]] ; then
                echo "UCP is all healty, good to move on!"
                break
            else
                echo "Not all healthy, rest and try again.."
                if [[ $n -eq 20 ]] ; then
                    echo "UCP failed status check after $n tries. Aborting..."
                    exit 0
                fi
                sleep 60
                let n+=1
            fi

        done
    }
    checkUCP

    # Checking if DTR is already running. If it is , exit, if it's not install it.
    if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$DTR_ELB_HOSTNAME/health) -ne 200 ]];
        # Installing first DTR replica
        # TODO: For upgrades, ensure that DTR isn't already installed
        then echo "Installing First DTR Replica..."
        sleep 30
        echo "Install DTR"
        date
        docker run --rm "$DTR_IMAGE" install --replica-https-port "$DTR_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$NODE_NAME" --dtr-external-url $DTR_ELB_HOSTNAME:443 --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --replica-id 000000000000
        echo "After running install via Docker"
        date
        sleep 30
        echo "Finished installing DTR"
    else
        exit 0
    fi

    # Checking if DTR is up
    checkDTR(){
        n=0
        until [ $n -gt 20 ];
        do
            if [[ $(curl --insecure --silent --output /dev/null --write-out '%{http_code}' https://$DTR_ELB_HOSTNAME/health) -eq 200 ]];
                then echo "Main DTR Replica is up! Starting DTR replica join process"
                break
            else
                if [[ $n -eq 20 ]];
                    then echo "DTR failed status check after $n tries. Aborting Installation..."
                    exit 0
                fi
                echo "Try #$n: checking DTR status..."
                sleep 30
                let n+=1
            fi
        done
    }
    checkDTR

    # Configuring DTR with S3
    echo "Configuring DTR with S3 Storage Backend..."
    if [[ $(curl --silent --output /dev/null --write-out '%{http_code}' -k -u $UCP_ADMIN_USER:$UCP_ADMIN_PASSWORD -X PUT "https://$DTR_ELB_HOSTNAME/api/v0/admin/settings/registry/simple" -d "{\"storage\":{\"delete\":{\"enabled\":true},\"maintenance\":{\"readonly\":{\"enabled\":false}},\"s3\":{\"rootdirectory\":\"\",\"region\":\"$REGION\",\"regionendpoint\":\"\",\"bucket\":\"$S3_BUCKET_NAME\",\"secure\": true}}}") -lt 300 ]];
        then echo "Successfully Configured DTR storage backend with S3"
    else
        echo "Failed to configure DTR storage backend with S3"
        # Additional Debugging Info:
        curl -v --write-out '%{http_code}' -k -u $UCP_ADMIN_USER:$UCP_ADMIN_PASSWORD -X PUT "https://$DTR_ELB_HOSTNAME/api/v0/admin/settings/registry/simple" -d "{\"storage\":{\"delete\":{\"enabled\":true},\"maintenance\":{\"readonly\":{\"enabled\":false}},\"s3\":{\"rootdirectory\":\"\",\"region\":\"$REGION\",\"regionendpoint\":\"\",\"bucket\":\"$S3_BUCKET_NAME\",\"secure\": true}}}"
    fi

    # Installing  DTR replicas
    # Check `docker node ls` for reachable non-leader managers and use them as ucp-node when joining DTR replicas one at a time.
    # TODO: Better error handling to ensure we only install it on nodes that don't have DTR already.
    for replica in $(docker node ls | grep Reachable | awk '{print $2}');
        do echo "Joining DTR replicas..." && sleep 30 && docker run --rm "$DTR_IMAGE" join --replica-https-port "$DTR_PORT" --ucp-url https://$UCP_ELB_HOSTNAME --ucp-node "$replica" --ucp-username "$UCP_ADMIN_USER" --ucp-password "$UCP_ADMIN_PASSWORD" --ucp-insecure-tls --existing-replica-id 000000000000
    done
    echo "Successfully joined DTR replicas!"

else
    echo "Not the Swarm leader. Exiting... "
fi

echo "Notify AWS that this manager node is ready"
cfn-signal --stack $STACK_NAME --resource $INSTANCE_NAME --region $REGION

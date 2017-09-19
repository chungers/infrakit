#!/bin/sh
# SUMMARY: Kill the docker daemon on a node in the swarm to test the aznodewatch feature
# LABELS: azure

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

# Get the hostname of the current node as to avoid taking it down
LOCAL_NODE=$(docker node inspect self --format "{{.Description.Hostname}}")
# Get the hostname of the node to take down
NODE_TO_KILL=$(docker node ls --format "{{.Hostname}}" -f "role=manager" | grep -v $LOCAL_NODE | head -1)
# Count the total number of managers started with and expected to be up at the end of the test
TOTAL_MANAGERS=$(docker info --format "{{.Swarm.Managers}}")

if [ $TOTAL_MANAGERS -lt 3 ]
then
    echoerr "Too few managers."
    exit 1
fi

# Send a docker command to use a privileged container and kill the remote node
ssh -o StrictHostKeyChecking=no -i $SSH_KEY docker@$NODE_TO_KILL "/usr/local/bin/docker run -i --privileged --pid=host debian nsenter -t 1 -m -n service docker stop" || echo "ssh ran"

# Wait to make sure that the node is removed
# If it's not taken down in 5 minutes the test will fail
echo "Waiting for node to come down."
sleep 1m
MANAGER_COUNT=$(docker info --format "{{.Swarm.Managers}}")
WAIT=20
until [ $MANAGER_COUNT -lt $TOTAL_MANAGERS ]
do
    if [ $WAIT -eq 0 ]
    then
        echoerr "Node not taken down"
        exit 1
    fi
    echo "Waiting for node to come down."
    sleep 1m
    MANAGER_COUNT=$(docker info --format "{{.Swarm.Managers}}")
    WAIT=$(( $WAIT - 1 ))
done

echo "Node succesfully killed and taken down."

# Wait to make sure that the node comes back up
# If the node does not come back up in 5 minutes the test will fail
MANAGER_COUNT=$(docker info --format "{{.Swarm.Managers}}")
WAIT=20
WAITED=0
until [ $MANAGER_COUNT -ge $TOTAL_MANAGERS ]
do
    if [ $WAIT -eq 0 ]
    then
        echoerr "New Swarm manager not added."
        exit 1
    fi
    echo "Waiting"
    sleep 1m
    MANAGER_COUNT=$(docker info --format "{{.Swarm.Managers}}")
    WAIT=$(( $WAIT - 1 ))
    WAITED=$(( $WAITED + 1))
done

echo "New manager node up."
echo "Checking that no extra managers are created."

# Ensure that no extra nodes are created
# Runs for as long as it took for a new manager to be brought up
until [ $WAITED -eq 0 ]
do
    if [ $MANAGER_COUNT -gt $TOTAL_MANAGERS ]
    then
	echoerr "Too many manager nodes created"
	exit 1
    fi
    echo "Ensuring no extra manager nodes created."
    sleep 1m
    MANAGER_COUNT=$(docker info --format "{{.Swarm.Managers}}")
    WAITED=$(( $WAITED - 1 ))
done

echo "New manager back up. No extra managers created."
exit 0

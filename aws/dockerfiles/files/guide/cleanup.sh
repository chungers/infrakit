#!/bin/bash
# this script cleans up and nodes that have been upgraded and no longer need to be in the swarm.
if [ "$NODE_TYPE" == "worker" ] ; then
    # this doesn't run on workers, only managers.
    exit 0
fi

# script runs via cron every 5 minutes, so all of them will start at the same time. Add a random
# delay so they don't step on each other when pulling items from the queue.
echo "Sleep for a short time (1-10 seconds). To prevent scripts from stepping on each other"
sleep $[ ( $RANDOM % 10 )  + 1 ]
echo "Finished sleep, lets get going."

# find any nodes that are marked as down, and remove from the
# DOWN_LIST=$(docker node inspect $(docker node ls -q) | jq -r '.[] | select(.Status.State == "down") | .ID')

MESSAGES=$(aws sqs receive-message --region $REGION --queue-url $CLEANUP_QUEUE  --max-number-of-messages 10 --wait-time-seconds 10 --visibility-timeout 5 )

COUNT=$(echo $MESSAGES | jq -r '.Messages | length')
echo "$COUNT messages"
for((i=0;i<$COUNT;i++)); do
    echo "Loop $i"
    BODY=$(echo $MESSAGES | jq -r '.Messages['${i}'].Body')
    echo "BODY=$BODY"
    RECEIPT=$(echo $MESSAGES | jq --raw-output '.Messages['${i}'] .ReceiptHandle')
    echo "RECEIPT=$RECEIPT"
    docker node rm $BODY
    RESULT=$?
    if [ $RESULT -eq 0 ]; then
        echo "We were able to remove node from swarm, delete message from queue"
        aws sqs delete-message --region $REGION --queue-url $CLEANUP_QUEUE --receipt-handle $RECEIPT
        echo "message deleted"
        SWARM_ID=$(docker info | grep ClusterID | cut -f2 -d: | sed -e 's/^[ \t]*//')
        buoy -event="node:remove" -swarm_id=$SWARM_ID -flavor=aws -node_id=$BODY
    else
        echo "We were not able to remove node from swarm, don't delete. RESULT=$RESULT"
    fi
done

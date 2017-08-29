#!/bin/sh
# Summary: Verify that you can deploy a service based on node constraints placed on label
# Labels:

# Description
# 1. Add a label to half the nodes
# 2. Verify that the nodes have the label
# 3. Deploy a service with a label constraint
# 4. Verify that only the nodes with the label have the service
# 5. Remove the labels
# 6. Verify that the labels have been removed

set -e
. "${RT_PROJECT_ROOT}/_lib/lib.sh"

REPLICAS=3
SERVICE_NAME="Top"$REPLICAS
KEY="other"
VALUE="true"

SERVICE_NODE_IDS_SORTED="service_node_ids_sorted.txt"
LABEL_NODE_IDS_SORTED="label_node_ids_sorted.txt"

clean_up(){
  docker service rm $SERVICE_NAME || echo "true"
  rm $SERVICE_NODE_IDS_SORTED || echo "true"
  rm $LABEL_NODE_IDS_SORTED || echo "true"
}

trap clean_up EXIT


# Select half the nodes to add a label to
NODE_COUNT=$(( $(docker info --format "{{json .Swarm.Nodes}}") / 2  ))

# If there is only one node
# then add the label to the node
# Test is more trivial in this instance because there are no other
# nodes where the service can be deployed
if [[ $NODE_COUNT -lt "1" ]];
        then NODE_COUNT=1
fi

# Gets half the nodes to add a label to 
LABEL_NODES=$(docker node ls -q  |  head -$NODE_COUNT)  

# Add a label to half of the nodes 
for node_id in $LABEL_NODES
do
        docker node update --label-add $KEY=$VALUE $node_id
        echo $node_id        
done


# Verify that the label has been added to the selected nodes
for node_id in $LABEL_NODES
do
    docker node inspect --format "{{.Spec.Labels.$KEY}}" $node_id | assert_contains $VALUE
    echo $node_id
done

echo "All the nodes have labels"


# Deploy a service with the label constraint
docker service create  --detach=false --name $SERVICE_NAME --replicas $REPLICAS --constraint "node.labels.${KEY} == ${VALUE}" alpine top 

echo "Create a service with a label constraint"

# Verify that only the nodes with the label have the service running

# Get all the nodes that have the service (getting host name)
# Then get the node id
SERVICES_NODES=$(docker service ps $SERVICE_NAME --format "{{.Node}}") 
for host_name in $SERVICE_NODES
do
    SERVICE_NODE_IDS=$SERVICE_NODE_IDS" "$(docker node inspect $host_name --format "{{.ID}}")
done

# Sort and remove duplicates so the files can be compared side by side
# Compare the nodes that have the labels with the nodes that have the service
echo $SERVICE_NODE_IDS|tr " " "\n"|sort -u|tr "\n" " " > $SERVICE_NODE_IDS_SORTED
echo $LABEL_NODES|tr " " "\n"|sort -u|tr "\n" " " > $LABEL_NODE_IDS_SORTED

while read service_node_id <&3 && read expected_node_id <&4; do
        assert_equals "Node with label has service" $service_node_id $expected_node_id
done 3<$SERVICE_NODE_IDS_SORTED 4<$LABEL_NODE_IDS_SORTED

echo "Removing all the labels"
# Remove all the labels
for node_id in $LABEL_NODES
do
        echo "Removing label: $KEY from $node_id"
        docker node update --label-rm "$KEY" $node_id 
done

# Verify that all the labels are removed
for node_id in $LABEL_NODES
do
    docker node inspect --format "{{.Spec.Labels.$KEY}}" $node_id | assert_contains "no value"
done

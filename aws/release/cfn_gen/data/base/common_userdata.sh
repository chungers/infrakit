
mkdir -p /var/lib/docker/editions
echo "$EXTERNAL_LB" > /var/lib/docker/editions/lb_name
echo "# hostname : ELB_name" >> /var/lib/docker/editions/elb.config
echo "127.0.0.1: $EXTERNAL_LB" >> /var/lib/docker/editions/elb.config
echo "localhost: $EXTERNAL_LB" >> /var/lib/docker/editions/elb.config
echo "default: $EXTERNAL_LB" >> /var/lib/docker/editions/elb.config

echo '{"experimental": '$DOCKER_EXPERIMENTAL', "labels":["os=linux", "region='$NODE_REGION'", "availability_zone='$NODE_AZ'", "instance_type='$INSTANCE_TYPE'", "node_type='$NODE_TYPE'"] ' > /etc/docker/daemon.json
if [ $ENABLE_CLOUDWATCH_LOGS == 'yes' ] ; then
   echo ', "log-driver": "awslogs", "log-opts": {"awslogs-group": "'$LOG_GROUP_NAME'", "tag": "{{.Name}}-{{.ID}}" }}' >> /etc/docker/daemon.json
else
   echo ' }' >> /etc/docker/daemon.json
fi

chown -R docker /home/docker/
chgrp -R docker /home/docker/
rc-service docker restart
sleep 5

# init-aws
docker run --label com.docker.editions.system --log-driver=json-file --restart=no -d -e DYNAMODB_TABLE=$DYNAMODB_TABLE -e NODE_TYPE=$NODE_TYPE -e REGION=$AWS_REGION -e STACK_NAME=$STACK_NAME -e STACK_ID="$STACK_ID" -e ACCOUNT_ID=$ACCOUNT_ID -e INSTANCE_NAME=$INSTANCE_NAME -e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION -e EDITION_ADDON=$EDITION_ADDON -e HAS_DDC=$HAS_DDC -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/init-aws:$DOCKER_FOR_IAAS_VERSION

# guide-aws
docker run --label com.docker.editions.system --log-driver=json-file --name=guide-aws --restart=always -d -e DYNAMODB_TABLE=$DYNAMODB_TABLE -e NODE_TYPE=$NODE_TYPE -e REGION=$AWS_REGION -e STACK_NAME=$STACK_NAME -e INSTANCE_NAME=$INSTANCE_NAME -e VPC_ID=$VPC_ID -e STACK_ID="$STACK_ID" -e ACCOUNT_ID=$ACCOUNT_ID -e SWARM_QUEUE="$SWARM_QUEUE" -e CLEANUP_QUEUE="$CLEANUP_QUEUE" -e RUN_VACUUM=$RUN_VACUUM -e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION -e EDITION_ADDON=$EDITION_ADDON -e HAS_DDC=$HAS_DDC -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/guide-aws:$DOCKER_FOR_IAAS_VERSION

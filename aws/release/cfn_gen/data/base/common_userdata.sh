echo "$EXTERNAL_LB" > /var/lib/docker/swarm/lb_name
echo "# hostname : ELB_name" >> /var/lib/docker/swarm/elb.config
echo "127.0.0.1: $EXTERNAL_LB" >> /var/lib/docker/swarm/elb.config
echo "localhost: $EXTERNAL_LB" >> /var/lib/docker/swarm/elb.config
echo "default: $EXTERNAL_LB" >> /var/lib/docker/swarm/elb.config

echo '{"experimental": true, "labels"=["os=linux", "region='$NODE_REGION'", "availability_zone='$NODE_AZ'", "instance_type='$INSTANCE_TYPE'", "node_type='$NODE_TYPE'"] ' > /etc/docker/daemon.json
if [ $ENABLE_CLOUDWATCH_LOGS == 'yes' ] ; then
   echo ', "log-driver": "awslogs", "log-opts": {"awslogs-group": "'$LOG_GROUP_NAME'", "tag": "{{.Name}}-{{.ID}}" }}' >> /etc/docker/daemon.json
else
   echo ' }' >> /etc/docker/daemon.json
fi

chown -R docker /home/docker/
chgrp -R docker /home/docker/
rc-service docker restart
sleep 5

docker run --label com.docker.editions.system --log-driver=json-file --restart=no -d -e DYNAMODB_TABLE=$DYNAMODB_TABLE -e NODE_TYPE=$NODE_TYPE -e REGION=$AWS_REGION -e STACK_NAME=$STACK_NAME -e STACK_ID="$STACK_ID" -e ACCOUNT_ID=$ACCOUNT_ID -e INSTANCE_NAME=$INSTANCE_NAME -e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log docker4x/init-aws:$DOCKER_FOR_IAAS_VERSION

docker run --label com.docker.editions.system --log-driver=json-file --name=guide-aws --restart=always -d -e DYNAMODB_TABLE=$DYNAMODB_TABLE -e NODE_TYPE=$NODE_TYPE -e REGION=$AWS_REGION -e STACK_NAME=$STACK_NAME -e INSTANCE_NAME=$INSTANCE_NAME -e VPC_ID=$VPC_ID -e STACK_ID="$STACK_ID" -e ACCOUNT_ID=$ACCOUNT_ID -e SWARM_QUEUE="$SWARM_QUEUE" -e CLEANUP_QUEUE="$CLEANUP_QUEUE" -e RUN_VACUUM=$RUN_VACUUM -e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker4x/guide-aws:$DOCKER_FOR_IAAS_VERSION

docker volume create --name sshkey

docker run --label com.docker.editions.system --log-driver=json-file -ti --rm --user root -v sshkey:/etc/ssh --entrypoint ssh-keygen docker4x/shell-aws:$DOCKER_FOR_IAAS_VERSION -A

docker run --label com.docker.editions.system --log-driver=json-file --name=shell-aws --restart=always -d -p 22:22 -v /home/docker/:/home/docker/ -v /var/run/docker.sock:/var/run/docker.sock -v /var/lib/docker/swarm/lb_name:/var/lib/docker/swarm/lb_name:ro -v /var/lib/docker/swarm/elb.config:/var/lib/docker/swarm/elb.config -v /usr/bin/docker:/usr/bin/docker -v /var/log:/var/log -v sshkey:/etc/ssh -v /etc/passwd:/etc/passwd:ro -v /etc/shadow:/etc/shadow:ro -v /etc/group:/etc/group:ro docker4x/shell-aws:$DOCKER_FOR_IAAS_VERSION

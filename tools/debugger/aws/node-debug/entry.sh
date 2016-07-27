#!/usr/bin/env bash
function h1 () {
    echo " ======== [" "$@" "] ========"
}

function h2 () {
    echo " ------ [" "$@" "] ----------"
}

function br () {
    echo ""
}

echo "##==================== Start Node Debugger =======================##"
h2 `date`

h1 "ec2 info"

h2 "hostname"
curl -s http://169.254.169.254/latest/meta-data/hostname
br

h2 "AMI"
curl -s http://169.254.169.254/latest/meta-data/ami-id
br

h2 "instance type"
curl -s http://169.254.169.254/latest/meta-data/instance-type
br

h2 "Instance ID"
curl -s http://169.254.169.254/latest/meta-data/instance-id
br

h2 "Local IP"
curl -s http://169.254.169.254/latest/meta-data/local-ipv4
br

h2 "Public IP"
curl -s http://169.254.169.254/latest/meta-data/public-ipv4
br

h2 "Availability Zone"
curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone
br

h2 "Security Group"
curl -s http://169.254.169.254/latest/meta-data/security-groups/
br

h2 "IAM Profile"
curl -s http://169.254.169.254/latest/meta-data/iam/info
br

h2 "Processes"
ps

h2 "Disk space"
df -h

h2 "Memory"
free -m

h2 "Kernel"
uname -a

h2 "uptime"
uptime

h1 "swarm node info"
h2 "docker Version"
docker version

h2 "docker info"
docker info

h2 "docker ps -a"
docker ps -a

h2 "docker volumes"
docker volume ls

h2 "docker network"
docker network ls

# only call these when we are a manager, or else we will get an error.
if [[ $(docker info | grep IsManager) == " IsManager: Yes" ]] ; then
    echo "=== This is a Manager Node === "
    h2 "Docker inspect self"
    docker node inspect self

    h2 "docker services"
    docker service ls

    for i in `docker service ls -q`; do
        h2 "service: $i"
        docker service inspect $i
    done

    h2 "docker inspect editions_controller"
    docker inspect editions_controller
    export image=$(docker inspect editions_controller | grep docker4x/controller | awk '{print $2}' | sed -e 's/,//g' -e 's/"//g')
    echo "Checking on ELB using controller image already loaded: $image"
    docker run $image elb describe $(grep default /var/lib/docker/swarm/elb.config | awk '{print $2}')
    if [ -e /var/lib/docker/swarm/elb.config ]; then
        echo "Local config file: /var/lib/docker/swarm/elb.config"
        cat /var/lib/docker/swarm/elb.config
    else
        echo "'/var/lib/docker/swarm/elb.config' does not exist"
    fi
else
    echo "=== This is a Worker Node === "
fi

h2 "tail -n 100 /var/log/docker.log"
tail -n 100 /var/log/docker.log


h2 `date`
echo "##==================== Finished Node Debugger ====================##"

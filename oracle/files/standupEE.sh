#!/bin/bash

DOCKER_EE_URL="https://storebits.docker.com/ee/oraclelinux/sub-e0fd4d3a-3933-4a89-a6b5-034024fe516f"

#Create the partition for direct-lvm
sudo parted -a optimal /dev/sdb mklabel gpt
sudo parted -a optimal /dev/sdb mkpart primary '0%' 256GB

#Install Docker EE
sudo sh -c "echo $DOCKER_EE_URL/oraclelinux > /etc/yum/vars/dockerurl"
sudo yum install -y yum-utils device-mapper-persistent-data lvm2
sudo yum-config-manager --add-repo $DOCKER_EE_URL/oraclelinux/docker-ee.repo
sudo yum -y install docker-ee

#Configure direct-lvm
sudo mkdir /etc/docker
sudo cat <<EOF >  /etc/docker/daemon.json
{
  "storage-driver": "devicemapper",
  "storage-opts": [
    "dm.directlvm_device=/dev/sdb1",
    "dm.thinp_percent=95",
    "dm.thinp_metapercent=1",
    "dm.thinp_autoextend_threshold=80",
    "dm.thinp_autoextend_percent=20",
    "dm.directlvm_device_force=true"
  ]
}
EOF

#Start Docker EE
sudo systemctl start docker

#Update host firewall
for i in 12376 12382 443 12385 12380 12381 12386 12379 2376 12384 12383 12387
do
	sudo firewall-cmd --zone=public --permanent --add-port=$i/tcp
done

sudo firewall-cmd --reload

#Install DTR... todo better way to fetch IP
sudo docker run --rm -it  \
	--name ucp \
	-v /var/run/docker.sock:/var/run/docker.sock \
	docker/ucp \
	install \
	--admin-username admin \
	--admin-password superDuperSecure \
	--host-address $(curl http://icanhazip.com)

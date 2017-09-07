#!/bin/sh

sudo echo "https://storebits.docker.com/ee/m/sub-01acd8fe-85c6-4195-bbf7-7df840a8a3e2/oraclelinux" > /etc/yum/vars/dockerurl
sudo yum-config-manager --add-repo https://storebits.docker.com/ee/m/sub-01acd8fe-85c6-4195-bbf7-7df840a8a3e2/oraclelinux/docker-ee.repo
sudo yum install -y yum-utils device-mapper-persistent-data lvm2
sudo yum makecache fast
sudo yum -y install docker-ee


# 0 Prerequisites
# 0.1 Environment Variables in Dockerfile
There's a series of environment variables for Terraform that are setup in the Dockerfile
```
 ENV TF_VAR_tenancy_ocid "ocid1.tenancy.oc1..aaaaaaaacgiaaochnv57rbet6pbbsft4hk665re4aamc6venf2ee3s7s5r7q"
 ENV TF_VAR_user_ocid "ocid1.user.oc1..aaaaaaaapon5xs337njn4f7pis7tsvgagraswmifxontzrmczd56gg2l5usq"
 ENV TF_VAR_fingerprint "be:31:65:ac:bd:8a:73:50:49:e3:93:5f:1b:62:b9:1a"
 ENV TF_VAR_private_key_path "/root/.ssh/id_rsa_oracle"
 ENV TF_VAR_region "us-phoenix-1"
 ENV TF_VAR_compartment_ocid "ocid1.tenancy.oc1..aaaaaaaacgiaaochnv57rbet6pbbsft4hk665re4aamc6venf2ee3s7s5r7q"
 ENV TF_VAR_ssh_public_key "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDtdOG+xN0qrvKghX3WCZo31igbdNRL9Tkn2G0qkWtbkmSmGDgNyfbduZU81+Di0UwbcWXyA2GRu/8kfR/HIg1LQldvhksDZ5CacbJiYAqqTFWeM7BiJco6TVcmH0948SnJUyZpsozkMZ3d4T2IaFN78ocuzM48LwI5Cq6rtMkmd8cbdIzQwwDf7F8f3+jqMDwAFkytBtdnJXS4HiwRmOxuJ50OcZKrpJhVSuzw0v0pu/zkP7lrlsqBHw21TelzB3oBoDwRGnvmoLtB0NADDID2r9ol25s63Ica4q8e1Szm0AfZbu37sVmErFspAb74RP6IekXPGJ1wVaoaJcOozJDJ obmc instance" 
 ENV TF_VAR_SubnetOCID "ocid1.subnet.oc1.phx.aaaaaaaajxguja4a7axmxu7c2vad4e5hs2grpassexywdybd5ceiklq3p7va"
```
# 0.2 SSH Keys
See: https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/credentials.htm
# 0.2.1 API Signing key
Store your API Signing keys here:
```
<editions>/oracle/keys/id_rsa_oracle.pub
<editions>/oracle/keys/id_rsa_oracle
```
# 0.2.2 Instance SSH
> We should consider dynamically generating these. 

Store your instance SSH keys here:
```
editions/oracle/keys/id_rsa_obmc_instance.pub
editions/oracle/keys/id_rsa_obmc_instance
```
# 0.3 UCP Config
editions/oracle/files/standupEE.sh needs the following arguments:

```
 DOCKER_EE_URL="https://storebits.docker.com/ee/oraclelinux/sub-e0fd4d3a-3933-4a89-a6b5-034024fe516f"
 admin user
 admin password
```
# 1 Create Node
terraform creates the node

# 2 Add volume
## 2.1 Terraform Volume Creation
terraform also creates the volume

## 2.2 Node Side
```
sudo iscsiadm -m node -o new -T iqn.2015-12.com.oracleiaas:f87f3554-bf5e-4399-a5a5-9a550f44bdef -p 169.254.2.2:3260

sudo iscsiadm -m node -o update -T iqn.2015-12.com.oracleiaas:f87f3554-bf5e-4399-a5a5-9a550f44bdef -n node.startup -v automatic

sudo iscsiadm -m node -T iqn.2015-12.com.oracleiaas:f87f3554-bf5e-4399-a5a5-9a550f44bdef -p 169.254.2.2:3260 -l

fdisk /dev/sdb
```
> Create new partition… don’t format

# 3 Install Docker EE

```
sudo sh -c 'echo "https://storebits.docker.com/ee/oraclelinux/sub-e0fd4d3a-3933-4a89-a6b5-034024fe516f/oraclelinux" > /etc/yum/vars/dockerurl'
```

```
#container-selinux doesn’t exist? 
sudo yum install -y yum-utils device-mapper-persistent-data lvm2
```

```
sudo yum-config-manager --add-repo https://storebits.docker.com/ee/oraclelinux/sub-e0fd4d3a-3933-4a89-a6b5-034024fe516f/oraclelinux/docker-ee.repo
```

```
sudo yum -y install docker-ee
```

# 4 Configure direct-lvm

```
mkdir /etc/docker
cat <<EOF >  /etc/docker/daemon.json
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
```

# 5 Start Docker EE

```
Systemctl start docker
```

# 6 Install UCP

# 6.1 Parameters?
 UCP_ADMIN_USER
 UCP_ADMIN_PASSWORD
 UCP_ADDITIONAL_ALIASES?

# 6.2 Update host firewall
```bash
for i in 12376 12382 443 12385 12380 12381 12386 12379 2376 12384 12383 12387; do sudo firewall-cmd --zone=public --permanent --add-port=$i/tcp; done
sudo firewall-cmd --reload
```

# 6.3 do the magic install
docker container run --rm -it --name ucp \
>   -v /var/run/docker.sock:/var/run/docker.sock \
>   docker/ucp:2.2.2 install \
>   --host-address <node-ip-address> \
>   --interactive


Rewrite firewall commands in terms of firewall-cmd?
```
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -D OUTPUT -m addrtype --dst-type LOCAL ! --dst 127.0.0.0/8 -j DOCKER' failed: iptables v1.4.21: Couldn't load target `DOCKER':No such file or directory#012#012Try `iptables -h' or 'iptables --help' for more information.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -D OUTPUT -m addrtype --dst-type LOCAL -j DOCKER' failed: iptables v1.4.21: Couldn't load target `DOCKER':No such file or directory#012#012Try `iptables -h' or 'iptables --help' for more information.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -D PREROUTING' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -D OUTPUT' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -F DOCKER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -X DOCKER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -F DOCKER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -X DOCKER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -F DOCKER-ISOLATION' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -X DOCKER-ISOLATION' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -n -L DOCKER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -n -L DOCKER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -n -L DOCKER-ISOLATION' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -C DOCKER-ISOLATION -j RETURN' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -n -L DOCKER-USER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -C DOCKER-USER -j RETURN' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -C FORWARD -j DOCKER-USER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost kernel: ctnetlink v0.93: registering with nfnetlink.
Sep 11 13:18:49 localhost dockerd: time="2017-09-11T13:18:49.263838798Z" level=info msg="Default bridge (docker0) is assigned with an IP address 172.17.0.0/16. Daemon option --bip can be used to set a preferred IP address"
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -C POSTROUTING -s 172.17.0.0/16 ! -o docker0 -j MASQUERADE' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -C DOCKER -i docker0 -j RETURN' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -D FORWARD -i docker0 -o docker0 -j DROP' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -C FORWARD -i docker0 -o docker0 -j ACCEPT' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -C FORWARD -i docker0 ! -o docker0 -j ACCEPT' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -C PREROUTING -m addrtype --dst-type LOCAL -j DOCKER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t nat -C OUTPUT -m addrtype --dst-type LOCAL -j DOCKER ! --dst 127.0.0.0/8' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -C FORWARD -o docker0 -j DOCKER' failed: iptables: No chain/target/match by that name.
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -C FORWARD -o docker0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT' failed: iptables: Bad rule (does a matching rule exist in that chain?).
Sep 11 13:18:49 localhost firewalld[1054]: WARNING: COMMAND_FAILED: '/usr/sbin/iptables -w2 -t filter -C FORWARD -j DOCKER-ISOLATION' failed: iptables: No chain/target/match by that name.
```



# 7 Install DTR on Workers




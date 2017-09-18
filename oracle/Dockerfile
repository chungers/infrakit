FROM hashicorp/terraform

RUN apk update \ 
 && apk add ca-certificates wget \
 && update-ca-certificates

RUN mkdir ~/archive \
 && cd ~/archive \
 && wget https://github.com/oracle/terraform-provider-baremetal/releases/download/v1.0.18/linux.tar.gz \
 && tar xzvf linux.tar.gz

RUN mkdir -p ~/.terraform.d/plugins
RUN mv ~/archive/linux_amd64/terraform-provider-baremetal ~/.terraform.d/plugins
RUN rm -rf ~/archive/

COPY files/ .

RUN  mkdir /root/.ssh/
COPY keys/id_rsa_oracle.pub /root/.ssh/
COPY keys/id_rsa_oracle /root/.ssh/
COPY keys/id_rsa_obmc_instance.pub /root/.ssh/
COPY keys/id_rsa_obmc_instance /root/.ssh/

### Authentication details
ENV TF_VAR_tenancy_ocid "ocid1.tenancy.oc1..aaaaaaaacgiaaochnv57rbet6pbbsft4hk665re4aamc6venf2ee3s7s5r7q"
ENV TF_VAR_user_ocid "ocid1.user.oc1..aaaaaaaapon5xs337njn4f7pis7tsvgagraswmifxontzrmczd56gg2l5usq"
ENV TF_VAR_fingerprint "be:31:65:ac:bd:8a:73:50:49:e3:93:5f:1b:62:b9:1a"
ENV TF_VAR_private_key_path "/root/.ssh/id_rsa_oracle"

### Region
ENV TF_VAR_region "us-phoenix-1"

### Compartment
ENV TF_VAR_compartment_ocid "ocid1.tenancy.oc1..aaaaaaaacgiaaochnv57rbet6pbbsft4hk665re4aamc6venf2ee3s7s5r7q"

### Public/private keys used on the instance
### I'm almost certainly doing this the wrong way
ENV TF_VAR_ssh_public_key "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDtdOG+xN0qrvKghX3WCZo31igbdNRL9Tkn2G0qkWtbkmSmGDgNyfbduZU81+Di0UwbcWXyA2GRu/8kfR/HIg1LQldvhksDZ5CacbJiYAqqTFWeM7BiJco6TVcmH0948SnJUyZpsozkMZ3d4T2IaFN78ocuzM48LwI5Cq6rtMkmd8cbdIzQwwDf7F8f3+jqMDwAFkytBtdnJXS4HiwRmOxuJ50OcZKrpJhVSuzw0v0pu/zkP7lrlsqBHw21TelzB3oBoDwRGnvmoLtB0NADDID2r9ol25s63Ica4q8e1Szm0AfZbu37sVmErFspAb74RP6IekXPGJ1wVaoaJcOozJDJ obmc instance"

# ENV TF_VAR_ssh_private_key '${file("~/.ssh/id_rsa_obmc_instance/"}'
## NOTE: These are not your api keys. More info on the right keys see
## https://docs.us-phoenix-1.oraclecloud.com/Content/Compute/Tasks/managingkeypairs.htm

## Specific to this example
### Choose a subnet that exists in the AD and compartment you are launching the instance in
ENV TF_VAR_SubnetOCID "ocid1.subnet.oc1.phx.aaaaaaaajxguja4a7axmxu7c2vad4e5hs2grpassexywdybd5ceiklq3p7va"
RUN terraform init

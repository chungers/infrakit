#Overview

#Use docker container with AWS CLI to save VHD URL to tmp directory
#Download the created template from the jenkins job
#Add the VHD URL into the new template

#deployAzure.sh script run within the azure-e2e test container arguments are passed in to specify worker count, resource group, location
#manager count, SPSECRET, SPID, template. This script deploys a swarm, and then runs another container which includes a set of tests

VHDVAR=$(shell cat ${PWD}/tmp/vhd_blob_url.out) 

#Set values or default values depending on what is passed in

INITIAL_TEMPLATE_FILE := azurePublic.json
FINAL_TEMPLATE_FILE := updatedTemplate.json
AZURE_TEMPLATE_CONTAINER := azure_arm_template
E2E_CONTAINER := azure_e2e_test


ifeq ($(VHD_PREFIX), )
	VHD_PREFIX :=  s3://docker-ci-editions/vhd
endif

ifeq ($(TEMPLATE_PREFIX), )
	TEMPLATE_PREFIX := https://editions-stage-us-east-1-150610-005505.s3.amazonaws.com
endif

ifeq ($(MOBY_COMMIT), )
	MOBY_COMMIT := 36ab369660c0ab1c83d0b5a82864ae5b93a5c096
endif

ifeq ($(BUILDNUMBER), )
	BUILDNUMBER := 289
endif

ifeq ($(CHANNEL), )
	CHANNEL := test
endif

#Added timestamp so you should never have duplicate resource group names
ifeq ($(RESOURCEGROUP), )
	RESOURCEGROUP := testing_$(BUILDNUMBER)_$(shell date +%s)  
endif

ifeq ($(MANCOUNT), )
	MANCOUNT := 1
endif

ifeq ($(WORKERCOUNT), )
	WORKERCOUNT := 1
endif

ifeq ($(LOC), )
	LOC := southcentralus
endif

ifeq ($(PORT), )
	PORT := 50000
endif

ifeq ($(TENANT), )
	TENANT := azuresubdocker.onmicrosoft.com
endif



#Add the VHD URL to the template, and then use the created template to deploy in azure 
# -c option (clean) can be omitted and is used to indicate whether the resource
#  group will be deleted at the conclusion of the end to end test
#
# azure_arm_template: container which contains python script that places the vhd url
# into the azure json template
#
# azure_e2e_test: container which deploys azure swarm and runs set of e2e tests

default: setup
	docker build . --tag $(E2E_CONTAINER) 
	docker run -ti \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v /usr/bin/docker:/usr/bin/docker \
	$(E2E_CONTAINER) \
	-n "$(TENANT)" \
	-p "$(PORT)" \
	-m "$(MANCOUNT)" \
	-w "$(WORKERCOUNT)" \
	-l "$(LOC)" \
	-r "$(RESOURCEGROUP)" \
	-t /usr/bin/"$(FINAL_TEMPLATE_FILE)" \
	-s "$(SPSECRET)" \
	-i "$(SPID)"  \
	-c
	$(MAKE) clean

#Get the VHD URL, and the latest template
assets: 
	docker run --rm -v ${PWD}/tmp:/tmp docker4x/awscli:latest s3 --no-sign-request cp $(VHD_PREFIX)/$(MOBY_COMMIT)/vhd_blob_url.out /tmp
	curl  $(TEMPLATE_PREFIX)/azure/$(CHANNEL)/$(MOBY_COMMIT)/$(BUILDNUMBER)/Docker.tmpl >> tmp/$(INITIAL_TEMPLATE_FILE)  

#Update the latest template with the VHD url
setup: checks assets
	docker build . -f Dockerfile.azuretemplate --tag $(AZURE_TEMPLATE_CONTAINER) 
	docker run --rm -v ${PWD}:/v $(AZURE_TEMPLATE_CONTAINER)  /v/tmp/$(INITIAL_TEMPLATE_FILE) $(VHDVAR)  /v/tmp/$(FINAL_TEMPLATE_FILE)

#Make sure the SPID and SPSECRET are set
checks:
ifeq ($(SPID), )
	$(error "SPID is not set")
endif
ifeq ($(SPSECRET), )
	$(error "SPSECRET is not set")
endif


#Delete the tmp files, folders, and docker images
clean:
	rm -f ${PWD}/tmp/*
	rmdir ${PWD}/tmp
	docker image rm -f docker4x/awscli:latest
	docker image rm -f $(E2E_CONTAINER) 
	docker image rm -f $(AZURE_TEMPLATE_CONTAINER) 

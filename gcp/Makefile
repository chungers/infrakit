# Copyright 2016 Docker Inc. All rights reserved.

CLOUDSDK_CORE_PROJECT ?= docker-for-gcp
EDITIONS_VERSION ?= latest
BUCKET ?= gs://docker-for-gcp-templates

DEPLOYMENT_NAME ?= docker
MANAGER_COUNT ?= 3
WORKER_COUNT ?= 1
ZONE ?= europe-west1-d
MANAGER_MACHINE_TYPE ?= g1-small
WORKER_MACHINE_TYPE ?= g1-small

GCLOUD_IMAGE_TAG ?= sha256:fb904276e8a902ccd9564989d9222bdfbe37ffcd7f9989ca7e24b4019a9b4b6b
GCLOUD = docker run --rm -ti \
	-v $(CURDIR)/templates:/templates \
	--volumes-from gcloud-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gcloud
DEPLOYMENTS = $(GCLOUD) deployment-manager deployments
GSUTIL = docker run --rm -ti \
  -v $(CURDIR)/templates:/templates \
	--volumes-from gsutil-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gsutil
FORCE_GSUTIL_AUTH ?= 1

.PHONY: all auth revoke create describe delete release integration-tests dockerfiles export-templates
.DEFAULT: all
all: auth create

auth: revoke
	docker run -it --name gcloud-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) true
	$(GCLOUD) init --skip-diagnostics

revoke:
	-docker rm gcloud-config 2>/dev/null

create:
	$(DEPLOYMENTS) create $(DEPLOYMENT_NAME) \
		--config /templates/swarm.jinja \
		--properties managerCount:$(MANAGER_COUNT),workerCount:$(WORKER_COUNT),zone:$(ZONE),managerMachineType:$(MANAGER_MACHINE_TYPE),workerMachineType:$(WORKER_MACHINE_TYPE)

update:
	$(DEPLOYMENTS) update $(DEPLOYMENT_NAME) \
		--config /templates/swarm.jinja \
		--properties managerCount:$(MANAGER_COUNT),workerCount:$(WORKER_COUNT),zone:$(ZONE),managerMachineType:$(MANAGER_MACHINE_TYPE),workerMachineType:$(WORKER_MACHINE_TYPE)

delete:
	-$(GCLOUD) compute instances delete --delete-disks=boot $(shell $(GCLOUD) compute instances list --filter='networkInterfaces[0].network ~ $(DEPLOYMENT_NAME)-network' --uri)
	$(DEPLOYMENTS) delete $(DEPLOYMENT_NAME)

release: dockerfiles export-templates

integration-tests:
	@echo "+ $@"
	@cd ${CURDIR}/integration; ./run.sh

dockerfiles:
	DOCKER_TAG=$(EDITIONS_VERSION) make -C dockerfiles all

export-templates:
ifeq "$(FORCE_GSUTIL_AUTH)" "1"
	-docker rm gsutil-config 2>/dev/null
	docker run -it --name gsutil-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) gcloud auth login --brief --project=$(CLOUDSDK_CORE_PROJECT)
endif
	$(GSUTIL) ls -b $(BUCKET) || gsutil mb $(BUCKET)
	$(GSUTIL) -m rsync -r /templates $(BUCKET)/$(EDITIONS_VERSION)
	$(GSUTIL) -m acl -r set public-read $(BUCKET)/$(EDITIONS_VERSION)
	$(GSUTIL) -m setmeta -h "Cache-Control:private, max-age=0, no-transform" -r $(BUCKET)/$(EDITIONS_VERSION)

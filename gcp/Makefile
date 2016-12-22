# Copyright 2016 Docker Inc. All rights reserved.

DEPLOYMENT_NAME ?= docker
MANAGER_COUNT ?= 3
WORKER_COUNT ?= 1
ZONE ?= europe-west1-d
MANAGER_MACHINE_TYPE ?= g1-small
WORKER_MACHINE_TYPE ?= g1-small

GCLOUD_IMAGE_TAG = sha256:44296dea85d3f8f016a9ac2d3225ac494c8cfc4593b0d3fb316eb26413e5846d
GCLOUD = docker run --rm -ti \
	-v $(CURDIR)/templates:/templates \
	--volumes-from gcloud-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gcloud
DEPLOYMENTS = $(GCLOUD) deployment-manager deployments

.PHONY: all auth revoke create describe delete
.DEFAULT: all
all: auth create

auth: revoke
	docker run -it --name gcloud-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) true
	$(GCLOUD) init --skip-diagnostics

revoke:
	docker rm gcloud-config 2>/dev/null || true

create:
	$(DEPLOYMENTS) create $(DEPLOYMENT_NAME) \
		--config /templates/swarm.jinja \
		--properties managerCount=$(MANAGER_COUNT),workerCount=$(MANAGER_COUNT),zone=$(ZONE),managerMachineType=$(MANAGER_MACHINE_TYPE),workerMachineType=$(WORKER_MACHINE_TYPE)

describe:
	$(DEPLOYMENTS) describe docker

delete:
	$(GCLOUD) compute instances delete $(shell $(GCLOUD) compute instances list --filter='networkInterfaces[0].network ~ $(DEPLOYMENT_NAME)-network' --uri) || true
	$(DEPLOYMENTS) delete $(DEPLOYMENT_NAME)

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
MANAGER_DISK_TYPE ?= pd-standard
WORKER_DISK_TYPE ?= pd-standard
MANAGER_DISK_SIZE ?= 100
WORKER_DISK_SIZE ?= 100
DEMO_MODE ?= true

GCLOUD_IMAGE_TAG ?= sha256:fb904276e8a902ccd9564989d9222bdfbe37ffcd7f9989ca7e24b4019a9b4b6b
GCLOUD = docker container run --rm -ti \
	-v $(CURDIR)/templates:/templates \
	--volumes-from gcloud-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gcloud
DEPLOYMENTS = $(GCLOUD) deployment-manager deployments
GSUTIL = docker container run --rm -ti \
  -v $(CURDIR)/templates:/templates \
	--volumes-from gsutil-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gsutil
FORCE_GSUTIL_AUTH ?= 1
TESTS ?= all

.PHONY: all auth revoke create describe delete release test dockerfiles export-templates
.DEFAULT: all

all: auth create

auth: revoke
	docker container run -it --name gcloud-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) true
	$(GCLOUD) init --skip-diagnostics

revoke:
	-docker container rm gcloud-config 2>/dev/null

create:
	$(DEPLOYMENTS) create $(DEPLOYMENT_NAME) \
		--config /templates/Docker.jinja \
		--properties managerCount:$(MANAGER_COUNT),workerCount:$(WORKER_COUNT),zone:$(ZONE),managerMachineType:$(MANAGER_MACHINE_TYPE),workerMachineType:$(WORKER_MACHINE_TYPE),managerDiskType:$(MANAGER_DISK_TYPE),workerDiskType:$(WORKER_DISK_TYPE),managerDiskSize:$(MANAGER_DISK_SIZE),workerDiskSize:$(WORKER_DISK_SIZE),demoMode:$(DEMO_MODE)

update:
	$(DEPLOYMENTS) update $(DEPLOYMENT_NAME) \
		--config /templates/Docker.jinja \
		--properties managerCount:$(MANAGER_COUNT),workerCount:$(WORKER_COUNT),zone:$(ZONE),managerMachineType:$(MANAGER_MACHINE_TYPE),workerMachineType:$(WORKER_MACHINE_TYPE),managerDiskType:$(MANAGER_DISK_TYPE),workerDiskType:$(WORKER_DISK_TYPE),managerDiskSize:$(MANAGER_DISK_SIZE),workerDiskSize:$(WORKER_DISK_SIZE),demoMode:$(DEMO_MODE)

delete:
	-$(GCLOUD) compute instances delete -q --delete-disks=boot $(shell $(GCLOUD) compute instances list --filter='networkInterfaces[0].network ~ $(DEPLOYMENT_NAME)-network' --uri)
	$(DEPLOYMENTS) delete -q $(DEPLOYMENT_NAME)

release: dockerfiles test export-templates

test:
	@cd ${CURDIR}/tests; docker-compose build && docker-compose run test $(TESTS)

dockerfiles:
	DOCKER_TAG=$(EDITIONS_VERSION) make -C dockerfiles all

export-templates:
ifeq "$(FORCE_GSUTIL_AUTH)" "1"
	-docker container rm gsutil-config 2>/dev/null
	docker container run -it --name gsutil-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) gcloud auth login --brief --project=$(CLOUDSDK_CORE_PROJECT)
endif
	$(GSUTIL) ls -b $(BUCKET) || gsutil mb $(BUCKET)
	$(GSUTIL) -m rsync -r /templates $(BUCKET)/$(EDITIONS_VERSION)
	$(GSUTIL) -m acl -r set public-read $(BUCKET)/$(EDITIONS_VERSION)
	$(GSUTIL) -m setmeta -h "Cache-Control:private, max-age=0, no-transform" -r $(BUCKET)/$(EDITIONS_VERSION)

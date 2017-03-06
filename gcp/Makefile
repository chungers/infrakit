# Copyright 2016 Docker Inc. All rights reserved.

BUILD_NUMBER ?= 0
DOCKER_REPO_OWNER ?= docker4x
CLOUDSDK_CORE_PROJECT ?= docker-for-gcp
EDITIONS_VERSION ?= dev
BUILDS_BUCKET ?= gs://docker-for-gcp-builds
MOBY_BUCKET ?= gs://docker-for-gcp-images
TEMPLATES_BUCKET ?= gs://docker-for-gcp-templates
TESTS ?= all
DOCKERFILE_DIRS := $(wildcard dockerfiles/*/)

# Stack properties
DEPLOYMENT_NAME ?= docker
MANAGER_COUNT ?= 3
WORKER_COUNT ?= 1
ZONE ?= europe-west1-d
MANAGER_MACHINE_TYPE ?= n1-standard-1
WORKER_MACHINE_TYPE ?= n1-standard-1
MANAGER_DISK_TYPE ?= pd-standard
WORKER_DISK_TYPE ?= pd-standard
MANAGER_DISK_SIZE ?= 100
WORKER_DISK_SIZE ?= 100
ENABLE_SYSTEM_PRUNE ?= true
REMOTE_SSH_ORIGIN ?= 0.0.0.0/0
DEMO_MODE ?= true

# GCloud
GCLOUD_IMAGE_TAG ?= sha256:fb904276e8a902ccd9564989d9222bdfbe37ffcd7f9989ca7e24b4019a9b4b6b
GCLOUD = docker container run --rm -ti \
	-v $(CURDIR)/build:/build \
	--volumes-from gcloud-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gcloud
DEPLOYMENTS = $(GCLOUD) deployment-manager deployments
GSUTIL = docker container run --rm -ti \
	-v $(CURDIR)/build:/build \
	--volumes-from gsutil-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gsutil
FORCE_GSUTIL_AUTH ?= 0

define last
$(lastword $(subst /, ,$1))
endef

.PHONY: all auth revoke create delete release test build build-buoy build-moby build-templates build-images gcloud-auth create-save-bucket save save-moby save-templates save-images release-moby release-templates release-images $(DOCKERFILE_DIRS)
.DEFAULT: all

all: auth create

auth: revoke
	docker container run -it --name gcloud-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) true
	$(GCLOUD) init --skip-diagnostics

revoke:
	-docker container rm gcloud-config 2>/dev/null

create:
	$(DEPLOYMENTS) create $(DEPLOYMENT_NAME) \
		--config /build/templates/Docker.jinja \
		--properties managerCount:$(MANAGER_COUNT),workerCount:$(WORKER_COUNT),zone:$(ZONE),managerMachineType:$(MANAGER_MACHINE_TYPE),workerMachineType:$(WORKER_MACHINE_TYPE),managerDiskType:$(MANAGER_DISK_TYPE),workerDiskType:$(WORKER_DISK_TYPE),managerDiskSize:$(MANAGER_DISK_SIZE),workerDiskSize:$(WORKER_DISK_SIZE),enableSystemPrune:${ENABLE_SYSTEM_PRUNE},demoMode:$(DEMO_MODE),remoteSshOrigin:$(REMOTE_SSH_ORIGIN)

update:
	$(DEPLOYMENTS) update $(DEPLOYMENT_NAME) \
		--config /build/templates/Docker.jinja \
		--properties managerCount:$(MANAGER_COUNT),workerCount:$(WORKER_COUNT),zone:$(ZONE),managerMachineType:$(MANAGER_MACHINE_TYPE),workerMachineType:$(WORKER_MACHINE_TYPE),managerDiskType:$(MANAGER_DISK_TYPE),workerDiskType:$(WORKER_DISK_TYPE),managerDiskSize:$(MANAGER_DISK_SIZE),workerDiskSize:$(WORKER_DISK_SIZE),enableSystemPrune:${ENABLE_SYSTEM_PRUNE},demoMode:$(DEMO_MODE),remoteSshOrigin:$(REMOTE_SSH_ORIGIN)

delete:
	-$(GCLOUD) compute instances delete -q --delete-disks=boot $(shell $(GCLOUD) compute instances list --filter='networkInterfaces[0].network ~ $(DEPLOYMENT_NAME)-network' --uri)
	$(DEPLOYMENTS) delete -q $(DEPLOYMENT_NAME)

test:
	@cd ${CURDIR}/tests; docker-compose build && docker-compose run test $(TESTS)

build-dir:
	mkdir -p build

build-buoy:
	make -C ../tools/buoy build

build-moby: build-images
	mkdir -p ../moby/packages/gcp/dockerimages
	cp ./build/images.tar ../moby/packages/gcp/dockerimages/
	make -C ../moby build/gcp/gce.img.tar.gz
	cp ../moby/build/gcp/gce.img.tar.gz ./build

build-templates: build-dir test
	rm -Rf ./build/templates/*
	cp -r ./templates ./build/

$(DOCKERFILE_DIRS):
	DOCKER_IMAGE=$(DOCKER_REPO_OWNER)/$(call last,$@)-gcp:$(BUILD_NUMBER) make -C dockerfiles/$(call last,$@)

build-images: build-dir build-buoy $(DOCKERFILE_DIRS)
	docker save \
		$(foreach DIR,$(DOCKERFILE_DIRS),$(DOCKER_REPO_OWNER)/$(call basename,$(call last,$(DIR)))-gcp:$(BUILD_NUMBER)) \
		--output ./build/images.tar

build: build-moby build-images build-templates

gcloud-auth:
ifeq "$(FORCE_GSUTIL_AUTH)" "1"
	-docker container rm gsutil-config 2>/dev/null
	docker container run -it --name gsutil-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) gcloud auth login --brief --project=$(CLOUDSDK_CORE_PROJECT)
endif

create-save-bucket: gcloud-auth
	$(GSUTIL) ls -b $(BUILDS_BUCKET) || gsutil mb $(BUILDS_BUCKET)

save-moby: create-save-bucket
	$(GSUTIL) -m cp /build/gce.img.tar.gz $(BUILDS_BUCKET)/$(BUILD_NUMBER)/

save-templates: create-save-bucket
	$(GSUTIL) -m rsync -r /build/templates $(BUILDS_BUCKET)/$(BUILD_NUMBER)/templates/

save-images: create-save-bucket
	$(GSUTIL) -m cp /build/images.tar $(BUILDS_BUCKET)/$(BUILD_NUMBER)/

save: save-moby save-images save-templates

retrieve-moby: build-dir gcloud-auth
	$(GSUTIL) cp $(BUILDS_BUCKET)/$(BUILD_NUMBER)/gce.img.tar.gz ./build/

retrieve-images: build-dir gcloud-auth
	$(GSUTIL) cp $(BUILDS_BUCKET)/$(BUILD_NUMBER)/images.tar ./build/

retrieve-templates: build-dir gcloud-auth
	mkdir -p ./build/templates
	rm -Rf ./build/templates/*
	$(GSUTIL) cp -r $(BUILDS_BUCKET)/$(BUILD_NUMBER)/templates ./build/

clean:
	rm -Rf ./build

retrieve: retrieve-images retrieve-moby retrieve-templates

release-moby: gcloud-auth
	echo Release Moby build:$(BUILD_NUMBER) to version:$(EDITIONS_VERSION)
	$(GSUTIL) cp ./build/gce.img.tar.gz $(MOBY_BUCKET)/$(EDITIONS_VERSION)/
	$(GSUTIL) -m setmeta -h "Cache-Control:private, max-age=0, no-transform" -r $(MOBY_BUCKET)/$(EDITIONS_VERSION)

release-templates: gcloud-auth
	echo Release Templates build:$(BUILD_NUMBER) to version:$(EDITIONS_VERSION)
	$(GSUTIL) cp -r ./build/templates/* $(TEMPLATES_BUCKET)/$(EDITIONS_VERSION)/
	sed -i.bak "s/VERSION = '.*'/VERSION = '$(EDITIONS_VERSION)'/" ./build/templates/Docker.jinja
	rm -f ./build/templates/Docker.jinja.bak
	$(GSUTIL) -m acl -r set public-read $(TEMPLATES_BUCKET)/$(EDITIONS_VERSION)
	$(GSUTIL) -m setmeta -h "Cache-Control:private, max-age=0, no-transform" -r $(TEMPLATES_BUCKET)/$(EDITIONS_VERSION)

release-images: gcloud-auth
	echo Release Images build:$(BUILD_NUMBER) to version:$(EDITIONS_VERSION)
	docker load -i ./build/images.tar
	for IMAGE in $(foreach DIR,$(DOCKERFILE_DIRS),$(call basename,$(call last,$(DIR)))); do \
		docker tag $(DOCKER_REPO_OWNER)/$$IMAGE-gcp:$(BUILD_NUMBER) $(DOCKER_REPO_OWNER)/$$IMAGE-gcp:$(EDITIONS_VERSION); \
		docker push $(DOCKER_REPO_OWNER)/$$IMAGE-gcp:$(EDITIONS_VERSION); \
	done

release: release-moby release-images release-templates

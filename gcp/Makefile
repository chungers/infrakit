# Copyright 2016 Docker Inc. All rights reserved.

BUILD_NUMBER ?= 0
NAMESPACE ?= docker4x
DOCKER_PUSH ?= false
CLOUDSDK_CORE_PROJECT ?= docker-for-gcp
EDITIONS_VERSION ?= dev
IMAGE_VERSION ?= $(shell whoami)-dev
BUILDS_BUCKET ?= gs://docker-for-gcp-builds
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
MANAGER_BOOT_DISK_SIZE ?= 10
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

.PHONY: all auth revoke create delete release test build build-buoy build-cloudstor build-moby build-templates build-images gcloud-auth create-save-bucket save save-moby save-templates $(DOCKERFILE_DIRS) integration-tests
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
		--properties managerCount:$(MANAGER_COUNT),workerCount:$(WORKER_COUNT),zone:$(ZONE),managerMachineType:$(MANAGER_MACHINE_TYPE),workerMachineType:$(WORKER_MACHINE_TYPE),managerDiskType:$(MANAGER_DISK_TYPE),workerDiskType:$(WORKER_DISK_TYPE),managerBootDiskSize:$(MANAGER_BOOT_DISK_SIZE),managerDiskSize:$(MANAGER_DISK_SIZE),workerDiskSize:$(WORKER_DISK_SIZE),enableSystemPrune:${ENABLE_SYSTEM_PRUNE},demoMode:$(DEMO_MODE),remoteSshOrigin:$(REMOTE_SSH_ORIGIN)

update:
	$(DEPLOYMENTS) update $(DEPLOYMENT_NAME) \
		--config /build/templates/Docker.jinja \
		--properties managerCount:$(MANAGER_COUNT),workerCount:$(WORKER_COUNT),zone:$(ZONE),managerMachineType:$(MANAGER_MACHINE_TYPE),workerMachineType:$(WORKER_MACHINE_TYPE),managerDiskType:$(MANAGER_DISK_TYPE),workerDiskType:$(WORKER_DISK_TYPE),managerBootDiskSize:$(MANAGER_BOOT_DISK_SIZE),managerDiskSize:$(MANAGER_DISK_SIZE),workerDiskSize:$(WORKER_DISK_SIZE),enableSystemPrune:${ENABLE_SYSTEM_PRUNE},demoMode:$(DEMO_MODE),remoteSshOrigin:$(REMOTE_SSH_ORIGIN)

delete:
	-$(GCLOUD) compute instances delete -q --delete-disks=boot $(shell $(GCLOUD) compute instances list --filter='networkInterfaces[0].network ~ $(DEPLOYMENT_NAME)-network' --uri)
	$(DEPLOYMENTS) delete -q $(DEPLOYMENT_NAME)

test:
	@cd ${CURDIR}/tests; docker-compose build && docker-compose run test $(TESTS)

build-dir:
	mkdir -p build

build-buoy:
	make -C ../tools/buoy build

build-cloudstor:
	if [ ! -e dockerfiles/cloudstor-rootfs.tar.gz ]; then          \
	    make -C ../tools/cloudstor build;                          \
	    cp ../tools/cloudstor/cloudstor-rootfs.tar.gz dockerfiles/;\
	fi
	cd dockerfiles;                                                              \
	tar zxvf cloudstor-rootfs.tar.gz;                                            \
	docker plugin rm -f "$(NAMESPACE)/cloudstor-gcp:$(IMAGE_VERSION)" || true;   \
	docker plugin create "$(NAMESPACE)/cloudstor-gcp:$(IMAGE_VERSION)" ./plugin; \
	rm -rf ./plugin;                                                             \
	if [ "$(DOCKER_PUSH)" = true ]; then                                            \
	    docker plugin push "$(NAMESPACE)/cloudstor-gcp:$(IMAGE_VERSION)";        \
	fi;                                                                          \
	cd ..

build-moby: build-images
	make -C ../moby clean
	mkdir -p ../moby/packages/gcp/dockerimages
	cp ./build/images.tar ../moby/packages/gcp/dockerimages/
	EDITIONS_VERSION=$(EDITIONS_VERSION) make -C ../moby build/gcp/gce.img.tar.gz
	cp ../moby/build/gcp/gce.img.tar.gz ./build

build-templates: build-dir test
	rm -Rf ./build/templates/*
	cp -r ./templates ./build/
	sed -i.bak "s/BUILD_NUMBER = '.*'/BUILD_NUMBER = '$(BUILD_NUMBER)'/" ./build/templates/Docker.jinja
	rm -f ./build/templates/Docker.jinja.bak

$(DOCKERFILE_DIRS):
	DOCKER_IMAGE=$(NAMESPACE)/$(call last,$@)-gcp make -C dockerfiles/$(call last,$@)

build-images: build-dir build-buoy $(DOCKERFILE_DIRS)
	docker save \
		$(foreach DIR,$(DOCKERFILE_DIRS),$(NAMESPACE)/$(call basename,$(call last,$(DIR)))-gcp) \
		--output ./build/images.tar

build: build-moby build-cloudstor build-images build-templates

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

save: save-moby save-templates

retrieve-moby: build-dir gcloud-auth
	$(GSUTIL) cp $(BUILDS_BUCKET)/$(BUILD_NUMBER)/gce.img.tar.gz ./build/

clean:
	rm -Rf ./build
	rm -f ./dockerfiles/cloudstor-rootfs.tar.gz

retrieve: build-dir gcloud-auth
	mkdir -p ./build
	$(GSUTIL) -m cp -r $(BUILDS_BUCKET)/$(BUILD_NUMBER)/* ./build/

# TODO: Should be on AWS
release: gcloud-auth
	echo Release build:$(BUILD_NUMBER) to version:$(EDITIONS_VERSION)
	sed -i.bak "s/VERSION = '.*'/VERSION = '$(EDITIONS_VERSION)'/" ./build/templates/Docker.jinja
	rm -f ./build/templates/Docker.jinja.bak
	$(GSUTIL) -m cp -a public-read -r ./build/templates/* $(TEMPLATES_BUCKET)/$(EDITIONS_VERSION)/
	$(GSUTIL) -m setmeta -h "Cache-Control:private, max-age=0, no-transform" -r $(TEMPLATES_BUCKET)/$(EDITIONS_VERSION)/

integration-tests:
	echo Run Integration tests from build $(BUILD_NUMBER)
	cd ${CURDIR}/integration; docker-compose build && docker-compose run e2e

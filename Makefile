ifeq ($(EDITIONS_TAG),)
	EDITIONS_TAG := ce-rc4
endif

ifeq ($(EDITIONS_DOCKER_VERSION),)
	EDITIONS_DOCKER_VERSION := 17.06.0
endif

ifeq ($(RELEASE),)
	RELEASE := false
endif

ifeq ($(RELEASE),false)
	ifdef JENKINS_BUILD
		DAY := $(shell date +"%m_%d_%Y")
		EDITIONS_TAG := $(EDITIONS_TAG)-$(DAY)
		# If not a release in Jenkins then use the head of docker/editions.moby
		MOBY_COMMIT := $(shell docker run --rm docker4x/mobyhead@sha256:823149afe4b4f2ee7ac45c7963193ae2274cb216760fb98631a3109aff05e4af)
	else
		EDITIONS_TAG := $(EDITIONS_TAG)-$(shell whoami)-dev
	endif
endif

# Check if MOBY_COMMIT has been defined via env
ifeq ($(MOBY_COMMIT),)
	MOBY_COMMIT := ef91e6873f2ac7fa3ff9b65c9fe2365fa594764a
endif

# Check if DOCKER_VERSION has been set
ifeq ($(DOCKER_VERSION),)
	DOCKER_VERSION := $(EDITIONS_DOCKER_VERSION)-$(EDITIONS_TAG)
endif

ifeq ($(DOCKER_PUSH),)
	DOCKER_PUSH := false
endif

# Push final image output to S3 bucket
ifeq ($(PUSH_BUILD_TO_S3),)
	PUSH_BUILD_TO_S3 := false
endif

# By default don't load all Docker Images into the AMI (shell is still loaded)
ifeq ($(LOAD_IMAGES),)
	LOAD_IMAGES := false
endif

# Check if CHANNEL has been defined
ifeq ($(CHANNEL),)
	CHANNEL := edge
endif

# Check if DOCKER_EXPERIMENTAL has been defined
ifeq ($(DOCKER_EXPERIMENTAL),)
	DOCKER_EXPERIMENTAL := false
endif

# Check if BUILD has been defined
ifeq ($(BUILD),)
	BUILD := 1
endif

## Jenkins ENV
ifeq ($(JENKINS_BUILD),)
	BUILD_NUMBER ?= manual-$(shell date +"%s")
	BUILD_TAG ?= jenkins-editions-$(BUILD_NUMBER)
endif

NAMESPACE := docker4x
AWS_TAG_VERSION := $(DOCKER_VERSION)-aws$(BUILD)
AZURE_TAG_VERSION := $(DOCKER_VERSION)-azure$(BUILD)
GCP_TAG_VERSION := $(DOCKER_VERSION)-gcp$(BUILD)
REGION := us-west-2
CHANNEL_DDC := alpha
CHANNEL_CLOUD := alpha
EDITION_ADDON := base

#### AWS Specific VARS
AMI_SRC_REGION := us-east-1
MAKE_AMI_PUBLIC := no
ACCOUNT_LIST_FILE_URL := https://s3.amazonaws.com/docker-for-aws/data/accounts.txt
DOCKER_AWS_ACCOUNT_URL := https://s3.amazonaws.com/docker-for-aws/data/docker_accounts.txt
AWS_AMI_LIST := https://s3.amazonaws.com/docker-for-aws/data/ami/$(DOCKER_VERSION)/ami_list.json

#### Azure Specific VARS
VHD_SKU := docker-ce
VHD_VERSION := 1.0.6
CONTAINER_NAME := mobylinux
# stage offer will have the -preview 
VHD_OFFER_ID := docker-ce
EE_VHD_SKU := docker-ee
EE_VHD_VERSION := 1.0.1
# stage offer will have the -preview 
EE_OFFER_ID := docker-ee


#### GCP Specific VARS
GCP_BUILD_NUMBER := 3


export

ROOT_DIR := ${CURDIR}

AZURE_TARGET_PATH := dist/azure/$(CHANNEL)/$(AZURE_TAG_VERSION)
AZURE_TARGET_TEMPLATE := $(AZURE_TARGET_PATH)/Docker.tmpl
AWS_TARGET_PATH := dist/aws/$(CHANNEL)/$(AWS_TAG_VERSION)
AWS_TARGET_TEMPLATE := $(AWS_TARGET_PATH)/Docker.tmpl
GCP_TARGET_PATH := dist/gcp/$(CHANNEL)/$(GCP_TAG_VERSION)

export

UNAME_S := $(shell uname -s)
SEDFLAGS := -i
ifeq ($(UNAME_S),Darwin)
	SEDFLAGS := -i ""
endif

.PHONY: moby tools tools/buoy tools/metaserver tools/cloudstor

## Release and Nightly rely on the AMI/VHD Jenkins Metadata to get the Editions & Docker version
release: 
	$(MAKE) aws-release
	$(MAKE) azure-release
	$(MAKE) gcp-release

nightly:
	$(MAKE) aws-nightly
	$(MAKE) azure-nightly

templates:
	$(MAKE) azure-template
	$(MAKE) aws-template
	$(MAKE) gcp-template

## Container images targets
dockerimages: clean tools
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) dockerimages-aws
	$(MAKE) dockerimages-azure
	$(MAKE) dockerimages-gcp

dockerimages-aws: tools
	@echo "\033[32m+ $@ - EDITIONS_VERSION? ${AWS_TAG_VERSION}\033[0m"
	$(MAKE) -C aws/dockerfiles EDITIONS_VERSION=$(AWS_TAG_VERSION)

dockerimages-azure: tools
	@echo "\033[32m+ $@ - EDITIONS_VERSION? ${AZURE_TAG_VERSION}\033[0m"
	$(MAKE) -C azure/dockerfiles EDITIONS_VERSION=$(AZURE_TAG_VERSION)

dockerimages-gcp: tools
	@echo "\033[32m+ $@ - EDITIONS_VERSION? ${GCP_TAG_VERSION}\033[0m"
	$(MAKE) -C gcp build-cloudstor build-images EDITIONS_VERSION=$(GCP_TAG_VERSION)

dockerimages-walinuxagent:
	@echo "\033[32m+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}\033[0m"
	$(MAKE) -C azure walinuxagent TAG="$(AZURE_TAG_VERSION)"

define build_cp_tool
	$(MAKE) -C tools/$(1)
	mkdir -p aws/dockerfiles/$(3)
	mkdir -p azure/dockerfiles/$(3)
	mkdir -p gcp/dockerfiles/$(3)
	cp tools/$(1)/$(2) aws/dockerfiles/$(3)
	cp tools/$(1)/$(2) azure/dockerfiles/$(3)
	if [ "$(2)" = "guide" ]; then \
		cp tools/$(1)/$(2) gcp/dockerfiles/$(3); \
	fi
endef

define clean_plugin_tool
	-rm -f aws/dockerfiles/cloudstor-rootfs.tar.gz
	-rm -f azure/dockerfiles/cloudstor-rootfs.tar.gz
	-rm -f gcp/dockerfiles/cloudstor-rootfs.tar.gz
	-rm -Rf gcp/dockerfiles/meta
	-rm -Rf gcp/dockerfiles/init
endef

## General tools targets
tools: tools/buoy/bin/buoy tools/metaserver/bin/metaserver tools/cloudstor/cloudstor-rootfs.tar.gz

tools/buoy/bin/buoy:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(call build_cp_tool,buoy,bin/buoy,guide/bin)
	$(call build_cp_tool,buoy,bin/buoy,init/bin)

tools/metaserver/bin/metaserver:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(call build_cp_tool,metaserver,bin/metaserver,meta/bin)

tools/cloudstor/cloudstor-rootfs.tar.gz:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(call build_cp_tool,cloudstor,cloudstor-rootfs.tar.gz,.)

tools/awscli/image:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) -C tools/awscli

## Moby targets
moby/cloud/azure/vhd_blob_url.out: moby
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	sed $(SEDFLAGS) 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(AZURE_TAG_VERSION)"/' moby/packages/azure/etc/init.d/azure 
	$(MAKE) -C moby uploadvhd

moby/cloud/aws/ami_id.out: moby
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	sed $(SEDFLAGS) 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(AWS_TAG_VERSION)"/' moby/packages/aws/etc/init.d/aws
	$(MAKE) -C moby ami

moby/build/gcp/gce.img.tar.gz: moby
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) -C moby gcp-upload

moby/build/aws/initrd.img:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) -C moby build/aws/initrd.img

moby/build/azure/initrd.img:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) -C moby build/azure/initrd.img

moby/alpine/cloud/aws/ami_list.json: moby/alpine/cloud/aws/ami_id.out
	$(MAKE) -C aws/release replicate

moby:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) -C moby all

## Azure targets 
azure-dev: dockerimages-azure azure/editions.json moby/cloud/azure/vhd_blob_url.out
	# Temporarily going to continue to use azure/editions.json until the
	# development workflow gets refactored to use only top-level Makefile
	# for "running" in addition to "compiling".
	# 
	# Until then, 'make dev' in azure/ dir with proper parameters is a good
	# way to boot the Azure template.

azure-release:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) -C azure/release EDITIONS_VERSION=$(AZURE_TAG_VERSION)

$(AZURE_TARGET_TEMPLATE):
	$(MAKE) -C azure/release template EDITIONS_VERSION=$(AZURE_TAG_VERSION)

azure-template:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	# "easy use" alias to generate latest version of template.
	$(MAKE) $(AZURE_TARGET_TEMPLATE)

azure/editions.json: azure-template
	cp $(AZURE_TARGET_TEMPLATE) azure/editions.json

azure-nightly:
	@echo "\033[32m+ $@\033[0m"
	$(MAKE) -C azure nightly

## AWS Targets
aws-release:
	@echo "\033[32m+ $@\033[0m"
	$(MAKE) -C aws release

$(AWS_TARGET_TEMPLATE):
	$(MAKE) -C aws template EDITIONS_VERSION=$(AWS_TAG_VERSION)

aws-template:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	# "easy use" alias to generate latest version of template.
	$(MAKE) $(AWS_TARGET_TEMPLATE)

aws-nightly:
	@echo "\033[32m+ $@\033[0m"
	$(MAKE) -C aws nightly

## GCP Targets

gcp-template:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) -C gcp build-templates BUILD_NUMBER=$(GCP_BUILD_NUMBER)

gcp-release:
	$(MAKE) -C gcp release BUILD_NUMBER=$(GCP_BUILD_NUMBER) EDITIONS_VERSION=$(GCP_TAG_VERSION)

## Golang targets
# Package list
PKGS_AND_MOCKS := $(shell go list ./... | grep -v /vendor)
PKGS := $(shell echo $(PKGS_AND_MOCKS) | tr ' ' '\n' | grep -v /mock$)

get-gomock:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	-go get github.com/golang/mock/gomock
	-go get github.com/golang/mock/mockgen

generate:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	@go generate -x $(PKGS_AND_MOCKS)

test:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	@go test -v github.com/docker/editions/pkg/loadbalancer

check:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"

clean:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION}\033[0m"
	$(MAKE) -C tools/buoy clean
	$(MAKE) -C tools/metaserver clean
	$(MAKE) -C tools/cloudstor clean
	$(MAKE) -C tools/swarm-exec clean
	$(MAKE) -C moby clean
	rm -rf dist/
	rm -f $(AWS_TARGET_PATH)/*.tar
	rm -f $(AZURE_TARGET_PATH)/*.tar
	rm -f moby/cloud/azure/vhd_blob_url.out
	rm -f moby/cloud/aws/ami_id.out
	rm -f moby/cloud/aws/ami_id_ee.out
	$(call clean_plugin_tool)

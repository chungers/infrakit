ifeq (${EDITIONS_TAG},)
	EDITIONS_TAG := ce-rc1
endif

ifeq (${EDITIONS_DOCKER_VERSION},)
	EDITIONS_DOCKER_VERSION := 17.05.0
endif

ifeq (${RELEASE},)
	RELEASE := 0
endif

ifeq ($(RELEASE),0)
	ifdef JENKINS_BUILD
		DAY := $(shell date +"%m_%d_%Y")
		EDITIONS_TAG := $(EDITIONS_TAG)-$(DAY)
	else
		EDITIONS_TAG := $(EDITIONS_TAG)-$(shell whoami)-dev
	endif
endif

ifeq (${EDITIONS_VERSION},)
	EDITIONS_VERSION := $(EDITIONS_DOCKER_VERSION)-$(EDITIONS_TAG)
endif

ifeq (${DOCKER_PUSH},)
	DOCKER_PUSH := 1
endif

# Push final image output to S3 bucket
ifeq (${PUSH_BUILD_TO_S3},)
	PUSH_BUILD_TO_S3 := false
endif

# By default don't load Docker Images into the AMI
ifeq (${LOAD_IMAGES},)
	LOAD_IMAGES := false
endif

# Check if CHANNEL has been defined
ifeq (${CHANNEL},)
	CHANNEL := edge
endif

# Check if DOCKER_EXPERIMENTAL has been defined
ifeq (${DOCKER_EXPERIMENTAL},)
	DOCKER_EXPERIMENTAL := 1
endif

# Check if BUILD has been defined
ifeq (${BUILD},)
	BUILD := 1
endif

NAMESPACE := docker4x
AWS_TAG_VERSION := $(EDITIONS_VERSION)-aws$(BUILD)
AZURE_TAG_VERSION := $(EDITIONS_VERSION)-azure$(BUILD)
GCP_TAG_VERSION := $(EDITIONS_VERSION)-gcp$(BUILD)
REGION := us-west-2
CHANNEL_DDC := alpha
CHANNEL_CLOUD := alpha
EDITION_ADDON := base

#### Azure Specific VARS
VHD_SKU := docker-ce
VHD_VERSION := 1.0.3
# stage offer will have the -preview 
VHD_OFFER_ID := docker-ce
EE_VHD_SKU := docker-ee
EE_VHD_VERSION := 1.0.0
# stage offer will have the -preview 
EE_OFFER_ID := docker-ee

export

ROOTDIR := $(shell pwd)

AZURE_TARGET_PATH := dist/azure/$(CHANNEL)/$(AZURE_TAG_VERSION)
AZURE_TARGET_TEMPLATE := $(AZURE_TARGET_PATH)/Docker.tmpl
AWS_TARGET_PATH := dist/aws/$(CHANNEL)/$(AWS_TAG_VERSION)
AWS_TARGET_TEMPLATE := $(AWS_TARGET_PATH)/Docker.tmpl

UNAME_S := $(shell uname -s)
SEDFLAGS := -i
ifeq ($(UNAME_S),Darwin)
	SEDFLAGS := -i ""
endif

.PHONY: moby tools tools/buoy tools/metaserver tools/cloudstor

release: moby/cloud/aws/ami_id.out moby/cloud/azure/vhd_blob_url.out dockerimages
	$(MAKE) -C aws/release AMI=$(shell cat moby/cloud/aws/ami_id.out)
	# VHD=$(shell cat moby/cloud/azure/vhd_blob_url.out)

## Container images targets
dockerimages: tools
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(MAKE) dockerimages-aws
	$(MAKE) dockerimages-azure

dockerimages-aws: tools
	$(MAKE) -C aws/dockerfiles

dockerimages-azure: tools
	$(MAKE) -C azure/dockerfiles

dockerimages-walinuxagent:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(MAKE) -C azure walinuxagent TAG="azure-v$(EDITIONS_VERSION)"

define build_cp_tool
	$(MAKE) -C tools/$(1)
	mkdir -p aws/dockerfiles/files/$(3)
	mkdir -p azure/dockerfiles/files/$(3)
	mkdir -p gcp/dockerfiles/files/$(3)
	cp tools/$(1)/$(2) aws/dockerfiles/files/$(3)
	cp tools/$(1)/$(2) azure/dockerfiles/files/$(3)
	cp tools/$(1)/$(2) gcp/dockerfiles/files/$(3)
endef

define clean_plugin_tool
	-rm -f aws/dockerfiles/files/cloudstor-rootfs.tar.gz
	-rm -f azure/dockerfiles/files/cloudstor-rootfs.tar.gz
	-rm -f gcp/dockerfiles/files/cloudstor-rootfs.tar.gz
endef

## General tools targets
tools: tools/buoy/bin/buoy tools/metaserver/bin/metaserver tools/cloudstor/cloudstor-rootfs.tar.gz

tools/buoy/bin/buoy:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(call build_cp_tool,buoy,bin/buoy,bin)

tools/metaserver/bin/metaserver:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(call build_cp_tool,metaserver,bin/metaserver,bin)

tools/cloudstor/cloudstor-rootfs.tar.gz:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(call build_cp_tool,cloudstor,cloudstor-rootfs.tar.gz,.)

tools/awscli/image:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(MAKE) -C tools/awscli

## Moby targets
moby/cloud/azure/vhd_blob_url.out: moby
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	sed $(SEDFLAGS) 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(AZURE_TAG_VERSION)"/' moby/packages/azure/etc/init.d/azure 
	$(MAKE) -C moby uploadvhd

moby/cloud/aws/ami_id.out: moby
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	sed $(SEDFLAGS) 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(AWS_TAG_VERSION)"/' moby/packages/aws/etc/init.d/aws
	$(MAKE) -C moby ami

moby/build/gcp/gce.img.tar.gz: moby
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(MAKE) -C moby gcp-upload

moby/build/aws/initrd.img:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(MAKE) -C moby build/aws/initrd.img

moby/build/azure/initrd.img:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(MAKE) -C moby build/azure/initrd.img

moby:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(MAKE) -C moby all

check:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"

clean:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
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

## Azure targets 
azure-dev: dockerimages-azure azure/editions.json moby/cloud/azure/vhd_blob_url.out
	# Temporarily going to continue to use azure/editions.json until the
	# development workflow gets refactored to use only top-level Makefile
	# for "running" in addition to "compiling".
	# 
	# Until then, 'make dev' in azure/ dir with proper parameters is a good
	# way to boot the Azure template.

azure-release:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	$(MAKE) -C azure/release EDITIONS_VERSION=$(AZURE_TAG_VERSION)

$(AZURE_TARGET_TEMPLATE):
	$(MAKE) -C azure/release template EDITIONS_VERSION=$(AZURE_TAG_VERSION)

azure-template:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	# "easy use" alias to generate latest version of template.
	$(MAKE) $(AZURE_TARGET_TEMPLATE)

azure/editions.json: azure-template
	cp $(AZURE_TARGET_TEMPLATE) azure/editions.json


## Golang targets
# Package list
PKGS_AND_MOCKS := $(shell go list ./... | grep -v /vendor)
PKGS := $(shell echo $(PKGS_AND_MOCKS) | tr ' ' '\n' | grep -v /mock$)

get-gomock:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	-go get github.com/golang/mock/gomock
	-go get github.com/golang/mock/mockgen

generate:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	@go generate -x $(PKGS_AND_MOCKS)

test:
	@echo "+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}"
	@go test -v github.com/docker/editions/pkg/loadbalancer

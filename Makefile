.PHONY: moby tools tools/buoy tools/metaserver

EDITIONS_TAG := beta19
EDITIONS_DOCKER_VERSION := 17.03.0-ce-rc1
EDITIONS_VERSION := $(EDITIONS_DOCKER_VERSION)-$(EDITIONS_TAG)
REGION := us-west-2
CHANNEL ?= beta
CHANNEL_DDC := alpha
CHANNEL_CLOUD := alpha
DOCKER_EXPERIMENTAL ?= 1
VHD_SKU := docker4azure
VHD_VERSION := 1.13.7
# stage offer will have the -preview 
OFFER_ID := docker4azure
CS_VHD_SKU := docker4azure-cs-1_12
CS_VHD_VERSION := 1.0.4
# stage offer will have the -preview 
CS_OFFER_ID := docker4azure-cs-preview
RELEASE ?= 0
MOBY_GIT_REMOTE := git@github.com:docker/moby
MOBY_GIT_REVISION := 1.13.x
LOAD_IMAGES := true

ifeq ($(RELEASE),0)
EDITIONS_VERSION := $(EDITIONS_VERSION)-$(shell whoami)-dev
endif

export

ROOTDIR := $(shell pwd)

AZURE_TARGET_PATH := dist/azure/$(CHANNEL)/$(EDITIONS_VERSION)
AZURE_TARGET_TEMPLATE := $(AZURE_TARGET_PATH)/Docker.tmpl
AWS_TARGET_PATH := dist/aws/$(CHANNEL)/$(EDITIONS_VERSION)
AWS_TARGET_TEMPLATE := $(AWS_TARGET_PATH)/Docker.tmpl

release: moby/cloud/aws/ami_id.out moby/cloud/azure/vhd_blob_url.out dockerimages
	$(MAKE) -C aws/release AMI=$(shell cat moby/cloud/aws/ami_id.out)
	# VHD=$(shell cat moby/cloud/azure/vhd_blob_url.out)

dockerimages: tools
	dockerimages-aws
	dockerimages-azure
	
dockerimages-aws: tools
	$(MAKE) -C aws/dockerfiles

dockerimages-azure: tools
	$(MAKE) -C azure/dockerfiles

define build_cp_tool
	$(MAKE) -C tools/$(1)
	mkdir -p aws/dockerfiles/files/bin || true
	mkdir -p azure/dockerfiles/files/bin || true
	mkdir -p gcp/dockerfiles/guide/bin || true
	cp tools/$(1)/bin/$(1) aws/dockerfiles/files/bin/$(1)
	cp tools/$(1)/bin/$(1) azure/dockerfiles/files/bin/$(1)
	cp tools/$(1)/bin/$(1) gcp/dockerfiles/guide/bin/$(1)
endef

tools: tools/buoy/bin/buoy tools/metaserver/bin/metaserver

tools/buoy/bin/buoy:
	@echo "+ $@"
	$(call build_cp_tool,buoy)

tools/metaserver/bin/metaserver:
	$(call build_cp_tool,metaserver)

moby/cloud/azure/vhd_blob_url.out: moby
	sed -i 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(EDITIONS_VERSION)"/' moby/packages/azure/etc/init.d/azure 
	sed -i 's/export DOCKER_FOR_IAAS_VERSION_DIGEST=".*"/export DOCKER_FOR_IAAS_VERSION_DIGEST="$(shell cat azure/dockerfiles/walinuxagent/sha256.out)"/' moby/packages/azure/etc/init.d/azure 
	LOAD_IMAGES=false $(MAKE) -C moby uploadvhd

moby/cloud/aws/ami_id.out: moby
	sed -i 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(EDITIONS_VERSION)"/' moby/packages/aws/etc/init.d/aws
	TAG_KEY=$(EDITIONS_VERSION) $(MAKE) -C moby ami

moby/build/aws/initrd.img:
	$(MAKE) -C moby build/aws/initrd.img

moby/build/azure/initrd.img:
	$(MAKE) -C moby build/azure/initrd.img

moby:
	$(MAKE) -C moby all

clean:
	$(MAKE) -C tools/buoy clean
	$(MAKE) -C tools/metaserver clean
	$(MAKE) -C moby clean
	rm -f $(AWS_TARGET_PATH)/*.tar
	rm -f $(AZURE_TARGET_PATH)/*.tar
	rm -f moby/cloud/azure/vhd_blob_url.out
	rm -f moby/cloud/aws/ami_id.out

azure-dev: dockerimages-azure azure/editions.json moby/cloud/azure/vhd_blob_url.out
	# Temporarily going to continue to use azure/editions.json until the
	# development workflow gets refactored to use only top-level Makefile
	# for "running" in addition to "compiling".
	# 
	# Until then, 'make dev' in azure/ dir with proper parameters is a good
	# way to boot the Azure template.

azure-release:
	$(MAKE) -C azure/release

$(AZURE_TARGET_TEMPLATE):
	$(MAKE) -C azure/release template

azure-template:
	# "easy use" alias to generate latest version of template.
	$(MAKE) $(AZURE_TARGET_TEMPLATE)

azure/editions.json: azure-template
	cp $(AZURE_TARGET_TEMPLATE) azure/editions.json

# Package list
PKGS_AND_MOCKS := $(shell go list ./... | grep -v /vendor)
PKGS := $(shell echo $(PKGS_AND_MOCKS) | tr ' ' '\n' | grep -v /mock$)

get-gomock:
	@echo "+ $@"
	-go get github.com/golang/mock/gomock
	-go get github.com/golang/mock/mockgen

generate:
	@echo "+ $@"
	@go generate -x $(PKGS_AND_MOCKS)

test:
	@echo "+ $@"
	@go test -v github.com/docker/editions/pkg/loadbalancer

EDITIONS_TAG := ga-2
EDITIONS_DOCKER_VERSION := 1.13.1
EDITIONS_VERSION := $(EDITIONS_DOCKER_VERSION)-$(EDITIONS_TAG)
REGION := us-west-1
CHANNEL := stable
CHANNEL_DDC := alpha
CHANNEL_CLOUD := alpha
DOCKER_EXPERIMENTAL := 1
VHD_SKU := docker4azure
VHD_VERSION := 1.13.7
# stage offer will have the -preview 
OFFER_ID := docker4azure
CS_VHD_SKU := docker4azure-cs-1_12
CS_VHD_VERSION := 1.0.4
# stage offer will have the -preview 
CS_OFFER_ID := docker4azure-cs-preview
RELEASE := 0
MOBY_GIT_REMOTE := git@github.com:docker/moby
MOBY_GIT_REVISION := 1.13.x

ifeq ($(RELEASE),0)
EDITIONS_VERSION := $(EDITIONS_VERSION)-$(shell whoami)-dev
endif

export

AZURE_TARGET_TEMPLATE := dist/azure/$(CHANNEL)/azure-v$(EDITIONS_VERSION).json

release: moby/alpine/cloud/aws/ami_id.out moby/alpine/cloud/azure/vhd_blob_url.out dockerimages
	$(MAKE) -C aws/release AMI=$(shell cat moby/alpine/cloud/aws/ami_id.out)
	# VHD=$(shell cat moby/alpine/cloud/azure/vhd_blob_url.out)

dockerimages: tools/buoy
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

tools: tools/buoy tools/metaserver

tools/buoy:
	$(call build_cp_tool,buoy)

tools/metaserver:
	$(call build_cp_tool,metaserver)

moby/alpine/cloud/azure/vhd_blob_url.out: moby
	sed -i 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(EDITIONS_VERSION)"/' moby/alpine/packages/azure/etc/init.d/azure 
	sed -i 's/export DOCKER_FOR_IAAS_VERSION_DIGEST=".*"/export DOCKER_FOR_IAAS_VERSION_DIGEST="$(shell cat azure/dockerfiles/walinuxagent/sha256.out)"/' moby/alpine/packages/azure/etc/init.d/azure 
	git -C moby commit -asm "Bump Azure version to $(EDITIONS_VERSION)"
	$(MAKE) -C moby/alpine uploadvhd

moby/alpine/cloud/aws/ami_id.out: moby
	TAG_KEY=$(EDITIONS_VERSION) $(MAKE) -C moby/alpine ami

moby:
	git clone $(MOBY_GIT_REMOTE) moby
	git -C moby checkout $(MOBY_GIT_REVISION)

clean:
	$(MAKE) -C tools/buoy clean
	$(MAKE) -C tools/metaserver clean
	rm -rf moby

azure-dev: dockerimages-azure azure/editions.json moby/alpine/cloud/azure/vhd_blob_url.out
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

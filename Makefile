EDITIONS_TAG := beta12
EDITIONS_DOCKER_VERSION := 1.13.0-rc2
EDITIONS_VERSION := $(EDITIONS_DOCKER_VERSION)-$(EDITIONS_TAG)
REGION := us-west-1
CHANNEL := beta
CHANNEL_DDC := alpha
CHANNEL_CLOUD := alpha
DOCKER_EXPERIMENTAL := 1
VHD_SKU := docker4azure
VHD_VERSION := 1.13.1
# stage offer will have the -preview 
OFFER_ID := docker4azure
CS_VHD_SKU := docker4azure-cs-1_12
CS_VHD_VERSION := 1.0.4
# stage offer will have the -preview 
CS_OFFER_ID := docker4azure-cs-preview
export

release: moby/alpine/cloud/aws/ami_id.out moby/alpine/cloud/azure/vhd_blob_url.out dockerimages
	$(MAKE) -C aws/release AMI=$(shell cat moby/alpine/cloud/aws/ami_id.out)
	# VHD=$(shell cat moby/alpine/cloud/azure/vhd_blob_url.out)

dockerimages: buoy
	dockerimages-aws
	dockerimages-azure
	
dockerimages-aws:
	$(MAKE) -C aws/dockerfiles

dockerimages-azure:
	$(MAKE) -C azure/dockerfiles

buoy:
	$(MAKE) -C tools/buoy
	mkdir -p aws/dockerfiles/files/bin || true
	mkdir -p azure/dockerfiles/files/bin || true
	cp tools/buoy/bin/buoy aws/dockerfiles/files/bin/buoy
	cp tools/buoy/bin/buoy azure/dockerfiles/files/bin/buoy

moby/alpine/cloud/azure/vhd_blob_url.out: moby
	$(MAKE) -C moby/alpine azure

moby/alpine/cloud/aws/ami_id.out: moby
	TAG_KEY=$(EDITIONS_VERSION) $(MAKE) -C moby/alpine ami

moby:
	git clone git@github.com:docker/moby

clean:
	rm -rf moby

azure-release:
	$(MAKE) -C azure/release

azure-template:
	$(MAKE) -C azure/release template

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

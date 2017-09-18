include Makefile.variable

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
	#$(MAKE) oracle-release
	#$(MAKE) gcp-release

nightly:
	$(MAKE) aws-nightly
	$(MAKE) azure-nightly
	#$(MAKE) oracle-nightly

templates:
	$(MAKE) azure-template
	$(MAKE) aws-template
	#$(MAKE) oracle-template
	#$(MAKE) gcp-template

e2e:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) aws-e2e 
	$(MAKE) azure-e2e 
	#$(MAKE) oracle-e2e 


## Container images targets
dockerimages: deepclean tools
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) aws-dockerimages
	$(MAKE) azure-dockerimages
	$(MAKE) gcp-dockerimages
	$(MAKE) build-rtf


dockerimages-walinuxagent:
	@echo "\033[32m+ $@ - EDITIONS_VERSION: ${EDITIONS_VERSION}\033[0m"
	$(MAKE) -C azure walinuxagent TAG="$(AZURE_TAG_VERSION)"

define build_cp_tool
	$(MAKE) -C tools/$(1)
	mkdir -p aws/dockerfiles/$(3)
	mkdir -p azure/dockerfiles/$(3)
	#mkdir -p oracle/dockerfiles/$(3)
	cp tools/$(1)/$(2) aws/dockerfiles/$(3)
	cp tools/$(1)/$(2) azure/dockerfiles/$(3)
	cp tools/$(1)/$(2) oracle/dockerfiles/$(3)
	if [ "$(2)" = "bin/guide" ]; then \
		mkdir -p gcp/dockerfiles/$(3) \
		cp tools/$(1)/$(2) gcp/dockerfiles/$(3); \
	fi
endef

define clean_plugin_tool
	-rm -f aws/dockerfiles/cloudstor-rootfs.tar.gz
	-rm -f azure/dockerfiles/cloudstor-rootfs.tar.gz
	#-rm -f oracle/dockerfiles/cloudstor-rootfs.tar.gz
	-rm -f gcp/dockerfiles/cloudstor-rootfs.tar.gz
	-rm -f gcp/dockerfiles/guide/cloudstor-rootfs.tar.gz
	-rm -Rf gcp/dockerfiles/meta
	-rm -Rf gcp/dockerfiles/init
	-rm -Rf aws/dockerfiles/init/bin
	-rm -Rf aws/dockerfiles/guide/bin
	-rm -Rf aws/dockerfiles/meta/bin
	-rm -Rf azure/dockerfiles/init/bin
	-rm -Rf azure/dockerfiles/guide/bin
	#-rm -Rf azure/dockerfiles/meta/bin
	#-rm -Rf oracle/dockerfiles/init/bin
	#-rm -Rf oracle/dockerfiles/guide/bin
	#-rm -Rf oracle/dockerfiles/meta/bin
	# clean up common files copied
	-rm -Rf azure/dockerfiles/alb-controller/files/container/bin/
	-rm -f azure/dockerfiles/ddc-init/files/aztags.py
	-rm -f azure/dockerfiles/guide/files/aztags.py
	-rm -f azure/dockerfiles/init/files/aztags.py
	-rm -f azure/dockerfiles/logger/files/aztags.py
	-rm -f azure/dockerfiles/lookup/files/aztags.py
	-rm -f azure/dockerfiles/upgrade/files/aztags.py
endef

## General tools targets
tools: tools/buoy/bin/buoy tools/metaserver/bin/metaserver tools/cloudstor/cloudstor-rootfs.tar.gz

tools/buoy/bin/buoy:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(call build_cp_tool,buoy,bin/buoy,guide/bin)
	$(call build_cp_tool,buoy,bin/buoy,init/bin)

tools/metaserver/bin/metaserver:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(call build_cp_tool,metaserver,bin/metaserver,meta/bin)

tools/cloudstor/cloudstor-rootfs.tar.gz:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(call build_cp_tool,cloudstor,cloudstor-rootfs.tar.gz,.)

tools/awscli/image:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) -C tools/awscli

## Moby targets
moby/vhd: clean moby
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	sed $(SEDFLAGS) 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(AZURE_TAG_VERSION)"/' moby/packages/azure/etc/init.d/azure 
	$(MAKE) -C moby uploadvhd

moby/ami: clean moby
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	sed $(SEDFLAGS) 's/export DOCKER_FOR_IAAS_VERSION=".*"/export DOCKER_FOR_IAAS_VERSION="$(AWS_TAG_VERSION)"/' moby/packages/aws/etc/init.d/aws
	$(MAKE) -C moby ami

moby/bdi: clean moby
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) -C moby build/gcp/gce.img.tar.gz
	$(MAKE) -C gcp save-moby

moby/build/aws/initrd.img:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) -C moby build/aws/initrd.img

moby/build/azure/initrd.img:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) -C moby build/azure/initrd.img

moby/alpine/cloud/aws/ami_list.json: moby/alpine/cloud/aws/ami_id.out
	$(MAKE) -C aws/release replicate

moby:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) -C moby all

## Azure targets 
azure-dockerimages: tools
	@echo "\033[32m+ $@ - EDITIONS_VERSION? ${AZURE_TAG_VERSION}\033[0m"
	$(MAKE) -C azure/dockerfiles EDITIONS_VERSION=$(AZURE_TAG_VERSION)

azure-template:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	# "easy use" alias to generate latest version of template.
	$(MAKE) $(AZURE_TARGET_TEMPLATE)

azure-release:
	@echo "\033[32m+ $@ - Editions Commit: ${EDITIONS_COMMIT} \033[0m"
	$(MAKE) -C azure release EDITIONS_VERSION=$(AZURE_TAG_VERSION)

azure-dev: azure azure/editions.json moby/cloud/azure/vhd_blob_url.out-dockerimages
	# Temporarily going to continue to use azure/editions.json until the
	# development workflow gets refactored to use only top-level Makefile
	# for "running" in addition to "compiling".
	# 
	# Until then, 'make dev' in azure/ dir with proper parameters is a good
	# way to boot the Azure template.

$(AZURE_TARGET_TEMPLATE):
	$(MAKE) -C azure template EDITIONS_VERSION=$(AZURE_TAG_VERSION)

azure/editions.json: azure-template
	cp $(AZURE_TARGET_TEMPLATE) azure/editions.json

azure-nightly:
	@echo "\033[32m+ $@\033[0m"
	$(MAKE) -C azure nightly

azure-e2e:
	$(MAKE) -C azure/testing

## AWS Targets
aws-dockerimages: tools
	@echo "\033[32m+ $@ - EDITIONS_VERSION? ${AWS_TAG_VERSION}\033[0m"
	$(MAKE) -C aws/dockerfiles EDITIONS_VERSION=$(AWS_TAG_VERSION)

aws-template:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	# "easy use" alias to generate latest version of template.
	$(MAKE) $(AWS_TARGET_TEMPLATE)

aws-release:
	@echo "\033[32m+ $@ - Editions Commit: ${EDITIONS_COMMIT} \033[0m"
	$(MAKE) -C aws release

$(AWS_TARGET_TEMPLATE):
	$(MAKE) -C aws template EDITIONS_VERSION=$(AWS_TAG_VERSION)

aws-nightly:
	@echo "\033[32m+ $@\033[0m"
	$(MAKE) -C aws nightly

aws-e2e:
	$(MAKE) -C aws/test

## GCP Targets
gcp-dockerimages: tools
	@echo "\033[32m+ $@ - EDITIONS_VERSION? ${GCP_TAG_VERSION}\033[0m"
	$(MAKE) -C gcp build-cloudstor build-images

gcp-template:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) -C gcp build-templates

gcp-release: gcp-template
	$(MAKE) -C gcp save-templates
	$(MAKE) -C gcp release

## Oracle Targets
#  Definitely not copy-pasta from AWS, why do you ask?
#
oracle-dockerimages: tools
	@echo "\033[32m+ $@ - EDITIONS_VERSION? ${ORACLE_TAG_VERSION}\033[0m"
	$(MAKE) -C oracle/dockerfiles EDITIONS_VERSION=$(ORACLE_TAG_VERSION)

oracle-template:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	# "easy use" alias to generate latest version of template.
	$(MAKE) $(ORACLE_TARGET_TEMPLATE)

oracle-release:
	@echo "\033[32m+ $@ - Editions Commit: ${EDITIONS_COMMIT} \033[0m"
	$(MAKE) -C oracle release

$(ORACLE_TARGET_TEMPLATE):
	$(MAKE) -C oracle template EDITIONS_VERSION=$(ORACLE_TAG_VERSION)

oracle-nightly:
	@echo "\033[32m+ $@\033[0m"
	$(MAKE) -C oracle nightly

oracle-e2e:
	$(MAKE) -C oracle/test

## Golang targets
# Package list
PKGS_AND_MOCKS := $(shell go list ./... | grep -v /vendor)
PKGS := $(shell echo $(PKGS_AND_MOCKS) | tr ' ' '\n' | grep -v /mock$)

get-gomock:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	-go get github.com/golang/mock/gomock
	-go get github.com/golang/mock/mockgen

generate:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	@go generate -x $(PKGS_AND_MOCKS)

test:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	@go test -v github.com/docker/editions/pkg/loadbalancer

build-rtf:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} \033[0m"
	$(MAKE) -C tests build

ci-test:
	@echo "\033[32m+ $@ - Build templates \033[0m"
	$(MAKE) aws-template TEST_JENKINS=true
	$(MAKE) azure-template TEST_JENKINS=true
	@echo "\033[32m+ $@ - Checking templates syntax \033[0m"
	@for i in $(shell find . -iname "*.tmpl"); do jq . "$$i" > /dev/null || exit 1; done
	@echo "\033[32m+ $@ - Checking docker image creation \033[0m"
	$(MAKE) aws-dockerimages
	$(MAKE) azure-dockerimages

check:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"

clean:
	@echo "\033[32m+ $@ - DOCKER_VERSION: ${DOCKER_VERSION} - CHANNEL: ${CHANNEL}\033[0m"
	$(MAKE) -C tools/buoy clean
	$(MAKE) -C tools/metaserver clean
	$(MAKE) -C tools/cloudstor clean
	$(MAKE) -C tools/swarm-exec clean
	$(MAKE) -C moby clean
	rm -f moby/cloud/azure/$(VHD_OUT)
	rm -f moby/cloud/aws/$(AMI_OUT)
	rm -f moby/cloud/gcp/$(BDI_OUT)
	#TODO Oracle image cleanup OVA?
	$(call clean_plugin_tool)

deepclean: clean
	-docker run --rm -v ${CURDIR}/dist/:/data -e UID=$(shell id -u) alpine sh -c 'chmod ugo+rw -R -v /data/; chown $${UID}:$${UID} -v -R /data/'
	rm -rf dist/
	rm -f $(AWS_TARGET_PATH)/*.tar
	rm -f $(AZURE_TARGET_PATH)/*.tar
	#TODO ORACLE_TARGET_PATH cleaning

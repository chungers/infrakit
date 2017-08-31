include ../Makefile.variable

## Env vars
FORCE_GSUTIL_AUTH ?= 0

MOBY_IMG_URL := s3://moby-ci-artifacts/ci.qa.aws.dckr.io/${MOBY_COMMIT}/${MOBY_IMG_NAME}
REGEXTRACT_IMAGE=docker4x/regextract@sha256:a843d793ba6b83d4bb7d68102ca2c7dba0f985330581224e9afeacdbbf76b8db
GCE_IMAGE=mobylinux/mkimage-gce:83158ebe303a970a3131c3d02ef2ba00ca5e4a14@sha256:51c30fbf946ca9967688a411e7bf69769a22987cd47ad03428491dd1556fcf63
QEMU_IMAGE=mobylinux/qemu:0fb8c648e8ed9ef6b1ec449587aeab6c53872744@sha256:606f30d815102e73bc01c07915dc0d5f153b0252c63f5f0ed1e39621ec656eb5
TAR2INITRD_IMAGE=mobylinux/tar2initrd:d5711601eb5b89de0f052d87365e18388ff3f1b5@sha256:58d377e65845f91400e173ce9fca93462f2f237947eef2b0d2c17bb4f2da5ee8
TARTAR2INITRD_IMAGE=mobylinux/tartar2initrd:d56cde1558e3080e59a32e3cd7c7141baa601811@sha256:e1ad4522ff906d339da5f250b9ef6bffa5a70b4dec7d2cf7f7dbd0447b79352f
GCLOUD_IMAGE_TAG ?= sha256:fb904276e8a902ccd9564989d9222bdfbe37ffcd7f9989ca7e24b4019a9b4b6b

LOG_DIR := build

export

define push_to_s3
	@echo "+ Pushing ${3} to ${1} with type ${2} - ${3}"
	# Create versioned build
	docker run --rm \
		-v $(CURDIR)/$(LOG_DIR)/:/tmp \
		-e AWS_ACCESS_KEY_ID=${AWS_PUSH_ACCESS_KEY_ID} \
		-e AWS_SECRET_ACCESS_KEY=${AWS_PUSH_SECRET_ACCESS_KEY} \
		docker4x/awscli:latest s3 cp \
		/tmp/$(3) \
		s3://docker-ci-editions/$(2)/${EDITIONS_COMMIT}/$(BUILD_NUMBER)/$(3) \
		--acl public-read \
		--metadata editions_version=$(4),arch=$(1),channel=${CHANNEL},docker_version=${DOCKER_VERSION},moby_commit=${MOBY_COMMIT},editions_commit=$(EDITIONS_COMMIT),jenkins_build=$(BUILD_NUMBER)
	# Update latest for that commit
		docker run --rm \
		-v $(CURDIR)/$(LOG_DIR)/:/tmp \
		-e AWS_ACCESS_KEY_ID=${AWS_PUSH_ACCESS_KEY_ID} \
		-e AWS_SECRET_ACCESS_KEY=${AWS_PUSH_SECRET_ACCESS_KEY} \
		docker4x/awscli:latest s3 cp \
		/tmp/$(3) \
		s3://docker-ci-editions/$(2)/${EDITIONS_COMMIT}/$(3) \
		--acl public-read \
		--metadata editions_version=$(4),arch=$(1),channel=${CHANNEL},docker_version=${DOCKER_VERSION},moby_commit=${MOBY_COMMIT},editions_commit=$(EDITIONS_COMMIT),jenkins_build=$(BUILD_NUMBER)
	# Update the latest endpoint for nightly and tests
	@if [ "${EE}" = "no" ]; then \
		docker run --rm \
			-v $(CURDIR)/$(LOG_DIR)/:/tmp \
			-e AWS_ACCESS_KEY_ID=${AWS_PUSH_ACCESS_KEY_ID} \
			-e AWS_SECRET_ACCESS_KEY=${AWS_PUSH_SECRET_ACCESS_KEY} \
			docker4x/awscli:latest s3 cp \
			/tmp/$(3) \
			s3://docker-ci-editions/$(2)/$(3) \
			--acl public-read \
			--metadata editions_version=$(4),arch=$(1),channel=${CHANNEL},docker_version=${DOCKER_VERSION},moby_commit=${MOBY_COMMIT},editions_commit=$(EDITIONS_COMMIT),jenkins_build=$(BUILD_NUMBER); \
	fi
endef

all:
	@echo "\033[32m++ $@\033[0m"
	$(MAKE) src/initrd.img

# interactive versions need to use volume mounts
boot-qemu:
	@echo "\033[32m++ $@\033[0m"
	$(MAKE) src/initrd.img
	docker run -it --rm -v $(CURDIR)/src/initrd.img:/tmp/initrd.img -v $(CURDIR)/build/vmlinuz64:/tmp/vmlinuz64 $(QEMU_IMAGE)

src/initrd.img:
	@echo "\033[32m++ $@\033[0m"
	bash get.sh

gcp-upload: build/gcp/gce.img.tar.gz
	@echo "\033[32m++ $@\033[0m"
ifeq ($(FORCE_GSUTIL_AUTH),1)
	-docker rm gsutil-config 2>/dev/null
	docker run -it --name gsutil-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) gcloud auth login --brief --project=docker-for-gcp
endif
ifeq ($(JENKINS_BUILD),1)
	@echo "++ Creating auth volume for Jenkins"
	-docker rm gsutil-config 2>/dev/null
	docker run --name gsutil-config -v $(AUTH_FILE):/auth.json google/cloud-sdk@$(GCLOUD_IMAGE_TAG) gcloud auth activate-service-account --key-file /auth.json
endif
	docker run --rm \
		--volumes-from gsutil-config \
		-v $(CURDIR)/build/gcp/gce.img.tar.gz:/gce.img.tar.gz \
		google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
		gsutil cp -a public-read /gce.img.tar.gz gs://docker-for-gcp-images/${EDITIONS_COMMIT}/$(BUILD_NUMBER)/gce.img.tar.gz
		mkdir -p cloud/gcp
		echo "docker-for-gcp-images/${EDITIONS_COMMIT}/$(BUILD_NUMBER)/gce.img.tar.gz" >> ./$(LOG_DIR)/${BDI_OUT}
		$(MAKE) -C ../gcp save BUILD_NUMBER=$(GCP_BUILD_NUMBER) FORCE_GSUTIL_AUTH=0
ifeq ($(PUSH_BUILD_TO_S3),true)
	$(call push_to_s3,gcp,bdi,$(BDI_OUT),$(GCP_TAG_VERSION))
endif

ami:
	@echo "\033[32m++ $@\033[0m"
	$(MAKE) build/aws/initrd.img
	tar cf - \
	  cloud -C build/aws/ initrd.img -C ../ vmlinuz64 \
	  | \
	  docker build -t moby-ami:build -f cloud/Dockerfile.ami -
	# Clean previous AMI volume mounts
	$(MAKE) ami-clean-mount
	# The EBS device seems not to show up without mounting in /dev, even
	# with --privileged enabled.
	docker run \
		--rm \
		--privileged \
		-v /dev:/dev \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_ACCESS_KEY_ID \
		-e EDITIONS_VERSION=$(DOCKER_VERSION) \
		-e CHANNEL \
		moby-ami:build clean
	docker run \
		--rm \
		--privileged \
		-v /dev:/dev \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_ACCESS_KEY_ID \
		-e EDITIONS_VERSION=$(DOCKER_VERSION) \
		-e CHANNEL \
		-e MOBY_OUTPUT \
		-e LOAD_IMAGES \
		-e BUILD_TAG \
		moby-ami:build bake >  ./$(LOG_DIR)/$(AMI_OUT)
ifeq ($(PUSH_BUILD_TO_S3),true)
	$(call push_to_s3,aws,ami,$(AMI_OUT),$(AWS_TAG_VERSION))
endif

ami-clean-mount:
	@echo "\033[32m++ $@\033[0m"
	docker run \
		--rm \
		--privileged \
		-v /dev:/dev \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_ACCESS_KEY_ID \
		-e EDITIONS_VERSION=$(DOCKER_VERSION) \
		-e CHANNEL \
		-e MOBY_OUTPUT \
		-e DOCKER_BIN_URL \
		moby-ami:build clean-mount

uploadvhd:
	@echo "\033[32m++ $@\033[0m"
	$(MAKE) azure
	docker run \
		-i \
		-e VHD_SIZE \
		-e AZURE_STG_ACCOUNT_KEY \
		-e AZURE_STG_ACCOUNT_NAME \
		-e CONTAINER_NAME \
		-e BUILD_TAG \
		--log-driver none \
		--rm \
		-v vhdartifact:/tmp \
		moby-azure:build \
		uploadvhd >./$(LOG_DIR)/$(VHD_OUT)
ifeq ($(PUSH_BUILD_TO_S3),true)
	$(call push_to_s3,azure,vhd,$(VHD_OUT),$(AZURE_TAG_VERSION))
endif

azure:
	@echo "\033[32m++ $@\033[0m"
	$(MAKE) build/azure/initrd.img vhdartifact
	tar cf - \
	  cloud -C build/azure/ initrd.img -C ../ vmlinuz64 \
	  | \
	  docker build -t moby-azure:build -f cloud/Dockerfile.azure -
	tar cf - \
	  cloud/azure/raw2vhd.sh cloud/Dockerfile.raw2vhd \
	  | \
	  docker build -t moby-azure:raw2vhd -f cloud/Dockerfile.raw2vhd -
	# -v /dev:/dev needed in addition to --privileged due to creation of
	#  loopback device (mount namespace?)
	docker run \
		--rm \
		--privileged \
		--log-driver none \
		-v vhdartifact:/tmp \
		-v /dev:/dev \
		moby-azure:build \
		makeraw
	docker run \
		--rm \
		--log-driver none \
		-v vhdartifact:/tmp \
		moby-azure:raw2vhd
	docker run \
		--rm \
		-i \
		--log-driver none \
		-v vhdartifact:/tmp \
		moby-azure:build \
		tarout \
	| tar -xvf -

vhdartifact:
	@echo "\033[32m++ $@\033[0m"
	-docker volume create --name vhdartifact

aws.img:
	tar cf - -C packages/aws etc dockerimages -C ../.. | \
		docker run --rm --read-only --net=none --log-driver=none --tmpfs /tmp -i $(TAR2INITRD_IMAGE) > $@

build/aws/initrd.img: src/initrd.img aws.img
	cat $^ > $@
	rm aws.img

azure.img:
	@echo "\033[32m++ $@\033[0m"
	tar cf - -C packages/azure etc dockerimages -C ../.. | \
		docker run --rm --read-only --net=none --log-driver=none --tmpfs /tmp -i $(TAR2INITRD_IMAGE) > $@

build/azure/initrd.img: src/initrd.img azure.img
	cat $^ > $@
	rm azure.img

gce.img:
	@echo "\033[32m++ $@\033[0m"
	tar cf - -C packages/gcp etc dockerimages -C ../.. | \
		docker run --rm --read-only --net=none --log-driver=none --tmpfs /tmp -i $(TAR2INITRD_IMAGE) > $@

build/gcp/initrd.img: src/initrd.img gce.img
	cat $^ > $@
	rm gce.img

build/gcp/gce.img.tar.gz: build/gcp/initrd.img
	tar cf - -C build/gcp/ initrd.img -C ../ vmlinuz64 | docker run --rm --net=none --log-driver=none -i $(GCE_IMAGE) >$@
	cp build/gcp/gce.img.tar.gz ../gcp/build

clean:
	@echo "\033[32m++ $@\033[0m"
	-rm -rf build src packages/aws/dockerimages/ packages/azure/dockerimages/ packages/gcp/dockerimages/
	-rm -f *.img *.vhd *.iso *.tag mobylinux.efi gce.img.tar.gz
	-rm cloud/aws/$(AMI_OUT) cloud/azure/$(VHD_OUT) cloud/gcp/$(BDI_OUT)
 	-docker images -q moby-azure:build | xargs docker rmi -f
	-docker images -q moby-azure:raw2vhd | xargs docker rmi -f
	-docker volume rm vhdartifact

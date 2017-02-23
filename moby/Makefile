## Env vars
HASH := $(shell curl -f -s https://master.dockerproject.org/commit)
DAY := $(shell date +"%m_%d_%Y")
TAG_KEY ?= aws-nightly-$(DAY)-$(HASH)
CHANNEL ?= dev


## Moby specific commit
MOBY_IMG_COMMIT := b080a71704e707385e819747b0f2f4d66098292d
REGEXTRACT_IMAGE=docker4x/regextract@sha256:a843d793ba6b83d4bb7d68102ca2c7dba0f985330581224e9afeacdbbf76b8db
GCE_IMAGE=mobylinux/mkimage-gce:83158ebe303a970a3131c3d02ef2ba00ca5e4a14@sha256:51c30fbf946ca9967688a411e7bf69769a22987cd47ad03428491dd1556fcf63
QEMU_IMAGE=mobylinux/qemu:0fb8c648e8ed9ef6b1ec449587aeab6c53872744@sha256:606f30d815102e73bc01c07915dc0d5f153b0252c63f5f0ed1e39621ec656eb5
TAR2INITRD_IMAGE=mobylinux/tar2initrd:d5711601eb5b89de0f052d87365e18388ff3f1b5@sha256:58d377e65845f91400e173ce9fca93462f2f237947eef2b0d2c17bb4f2da5ee8
TARTAR2INITRD_IMAGE=mobylinux/tartar2initrd:d56cde1558e3080e59a32e3cd7c7141baa601811@sha256:e1ad4522ff906d339da5f250b9ef6bffa5a70b4dec7d2cf7f7dbd0447b79352f



all: src/initrd.img

# interactive versions need to use volume mounts
boot-qemu: 
	@echo "+ $@"
	$(MAKE) src/initrd.img
	docker run -it --rm -v $(CURDIR)/src/initrd.img:/tmp/initrd.img -v $(CURDIR)/build/vmlinuz64:/tmp/vmlinuz64 $(QEMU_IMAGE)

src/initrd.img: 
	@echo "+ $@"
	MOBY_IMG_COMMIT=$(MOBY_IMG_COMMIT) ./get.sh

gcp: gcp.img.tar.gz

gcp.img.tar.gz: build/gcp/initrd.img
	tar cf - -C build/gcp/ initrd.img -C ../ vmlinuz64 | docker run --rm --net=none --log-driver=none -i $(GCE_IMAGE) >$@

gcp-upload:
	@echo "+ $@"
ifeq ($(FORCE_GSUTIL_AUTH),1)
	-docker rm gsutil-moby-config 2>/dev/null
	docker run -it --name gsutil-moby-config google/cloud-sdk gcloud auth login --brief --project=docker-for-gcp
endif
	docker run --rm -ti \
		--volumes-from gsutil-moby-config \
		-v $(CURDIR)/gcp.img.tar.gz:/gcp.img.tar.gz \
		google/cloud-sdk \
		gsutil cp -a public-read /gcp.img.tar.gz gs://docker-for-gcp-images/latest/gcp.img.tar.gz

ami: 
	@echo "+ $@"
	$(MAKE) build/aws/initrd.img
	tar cf - \
	  cloud -C build/aws/ initrd.img -C ../ vmlinuz64 \
	  | \
	  docker build -t moby-ami:build -f cloud/Dockerfile.ami -
	# The EBS device seems not to show up without mounting in /dev, even
	# with --privileged enabled.
	docker run \
		--rm \
		--privileged \
		-v /dev:/dev \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_ACCESS_KEY_ID \
		-e TAG_KEY \
		-e TAG_KEY_PREV \
		-e CHANNEL \
		moby-ami:build clean
	docker run \
		--rm \
		--privileged \
		-v /dev:/dev \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_ACCESS_KEY_ID \
		-e TAG_KEY \
		-e TAG_KEY_PREV \
		-e CHANNEL \
		-e MOBY_OUTPUT \
		moby-ami:build bake > ./cloud/aws/ami_id.out

ami-clean-mount:
	@echo "+ $@"
	docker run \
		--rm \
		--privileged \
		-v /dev:/dev \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_ACCESS_KEY_ID \
		-e TAG_KEY \
		-e TAG_KEY_PREV \
		-e CHANNEL \
		-e MOBY_OUTPUT \
		-e DOCKER_BIN_URL \
		moby-ami:build clean-mount

uploadvhd: 
	@echo "+ $@"
	$(MAKE) azure
	docker run \
		-i \
		-e VHD_SIZE \
		-e AZURE_STG_ACCOUNT_KEY \
		-e AZURE_STG_ACCOUNT_NAME \
		-e CONTAINER_NAME \
		--log-driver none \
		--rm \
		-v vhdartifact:/tmp \
		moby-azure:build \
		uploadvhd >./cloud/azure/vhd_blob_url.out

azure: 
	@echo "+ $@"
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
	@echo "+ $@"
	# NB: Multiple 'docker volume create' with same name does not return
	# non-zero even though maybe it should.  The '|| true' is included as
	# future insurance.
	docker volume create --name vhdartifact || true


aws.img:
	tar cf - -C packages/aws etc dockerimages -C ../.. | \
		docker run --rm --read-only --net=none --log-driver=none --tmpfs /tmp -i $(TAR2INITRD_IMAGE) > $@

build/aws/initrd.img: src/initrd.img aws.img
	cat $^ > $@
	rm aws.img

azure.img:
	@echo "+ $@"
	tar cf - -C packages/azure etc -C ../.. | \
		docker run --rm --read-only --net=none --log-driver=none --tmpfs /tmp -i $(TAR2INITRD_IMAGE) > $@

build/azure/initrd.img: src/initrd.img azure.img
	cat $^ > $@
	rm azure.img

gcp.img:
	@echo "+ $@"
	tar cf - -C packages/gcp etc -C ../.. | \
		docker run --rm --read-only --net=none --log-driver=none --tmpfs /tmp -i $(TAR2INITRD_IMAGE) > $@

build/gcp/initrd.img: src/initrd.img gcp.img
	cat $^ > $@
	rm gcp.img

clean:
	@echo "+ $@"
	rm -rf build src packages/aws/dockerimages/ || true
	rm -f *.img *.vhd *.iso *.tag mobylinux.efi gcp.img.tar.gz || true
	docker images -q moby-azure:build | xargs docker rmi -f || true
	docker images -q moby-azure:raw2vhd | xargs docker rmi -f || true
	docker volume rm vhdartifact || true

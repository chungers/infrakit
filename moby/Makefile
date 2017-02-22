## Env vars
HASH := $(shell curl -f -s https://master.dockerproject.org/commit)
DAY := $(shell date +"%m_%d_%Y")
TAG_KEY := aws-nightly-$(DAY)-$(HASH)
CHANNEL := dev


## Moby specific commit
MOBY_IMG_COMMIT := 0bb61ead063c13e88b0603251ba218a8afe11757
REGEXTRACT_IMAGE=docker4x/regextract@sha256:a843d793ba6b83d4bb7d68102ca2c7dba0f985330581224e9afeacdbbf76b8db
GCE_IMAGE=mobylinux/mkimage-gce:83158ebe303a970a3131c3d02ef2ba00ca5e4a14@sha256:51c30fbf946ca9967688a411e7bf69769a22987cd47ad03428491dd1556fcf63
QEMU_IMAGE=mobylinux/qemu:0fb8c648e8ed9ef6b1ec449587aeab6c53872744@sha256:606f30d815102e73bc01c07915dc0d5f153b0252c63f5f0ed1e39621ec656eb5

all: build/initrd.img

# interactive versions need to use volume mounts
boot-qemu: build/initrd.img
	docker run -it --rm -v $(CURDIR)/build/initrd.img:/tmp/initrd.img -v $(CURDIR)/build/vmlinuz64:/tmp/vmlinuz64 $(QEMU_IMAGE)

build/initrd.img: 
	@echo "+ $@"
	MOBY_IMG_COMMIT=$(MOBY_IMG_COMMIT) ./get.sh

gce: gce.img.tar.gz

gce.img.tar.gz: build/initrd.img
	tar cf - -C build/ initrd.img vmlinuz64 | docker run --rm --net=none --log-driver=none -i $(GCE_IMAGE) >$@

gce-upload:
	@echo "+ $@"
ifeq ($(FORCE_GSUTIL_AUTH),1)
	-docker rm gsutil-moby-config 2>/dev/null
	docker run -it --name gsutil-moby-config google/cloud-sdk gcloud auth login --brief --project=docker-for-gcp
endif
	docker run --rm -ti \
		--volumes-from gsutil-moby-config \
		-v $(CURDIR)/gce.img.tar.gz:/gce.img.tar.gz \
		google/cloud-sdk \
		gsutil cp -a public-read /gce.img.tar.gz gs://docker-for-gcp-images/latest/gce.img.tar.gz

ami: 
	@echo "+ $@"
	$(MAKE) build/initrd.img
	tar cf - \
	  cloud -C build/ initrd.img vmlinuz64 \
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

# TODO(nathanleclaire): Migrate this to docker/editions repo.
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
	$(MAKE) build/initrd.img vhdartifact
	tar cf - \
	  cloud -C build/ initrd.img vmlinuz64 \
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

clean:
	@echo "+ $@"
	rm -f build || true
	rm -f *.img *.vhd *.iso *.tag mobylinux.efi gce.img.tar.gz || true
	docker images -q moby-azure:build | xargs docker rmi -f || true
	docker images -q moby-azure:raw2vhd | xargs docker rmi -f || true
	docker volume rm vhdartifact || true

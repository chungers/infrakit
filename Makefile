EDITIONS_TAG := beta5
EDITIONS_DOCKER_VERSION := 1.12.1
EDITIONS_VERSION := $(EDITIONS_DOCKER_VERSION)-$(EDITIONS_TAG)
REGION := us-west-1
CHANNEL := beta
CHANNEL_DDC := alpha
DOCKER_EXPERIMENTAL := 1
export

release: moby/alpine/cloud/aws/ami_id.out moby/alpine/cloud/azure/vhd_blob_url.out dockerimages
	$(MAKE) -C aws/release AMI=$(shell cat moby/alpine/cloud/aws/ami_id.out)
	# VHD=$(shell cat moby/alpine/cloud/azure/vhd_blob_url.out)

dockerimages: buoy
	$(MAKE) -C aws/dockerfiles
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

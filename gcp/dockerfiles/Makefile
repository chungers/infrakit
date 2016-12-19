EDITIONS_VERSION := 1.13.0-rc2-beta12

default: lb-controller

lb-controller:
	cd files/lb-controller/container && \
	VERSION="gcp-v$(EDITIONS_VERSION)" \
	DOCKER_TAG="gcp-v$(EDITIONS_VERSION)" \
	DOCKER_PUSH=true DOCKER_TAG_LATEST=false \
	make -k container

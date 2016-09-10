default: alb-controller
	VERSION="azure-v$(EDITIONS_VERSION)" ./build_and_push_all.sh

alb-controller:
	cd files/alb-controller/container && DOCKER_PUSH=true DOCKER_TAG_LATEST=true make -k container

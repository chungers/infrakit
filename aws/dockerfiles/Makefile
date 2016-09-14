default: elb-controller
	VERSION="aws-v$(EDITIONS_VERSION)" ./build_and_push_all.sh

elb-controller:
	cd files/elb-controller/container && DOCKER_PUSH=true DOCKER_TAG_LATEST=true make -k container

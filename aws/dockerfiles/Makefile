default: elb-controller
	VERSION="aws-v$(EDITIONS_VERSION)" ./build_and_push_all.sh

elb-controller:
	cd files/elb-controller/container && \
	VERSION="aws-v$(EDITIONS_VERSION)" \
	DOCKER_REPO="docker4x/l4controller-aws" \
	DOCKER_TAG="aws-v$(EDITIONS_VERSION)" \
	DOCKER_PUSH=true DOCKER_TAG_LATEST=true \
	make -k container

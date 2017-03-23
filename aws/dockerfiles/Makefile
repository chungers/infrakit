default: elb-controller
	VERSION="$(AWS_TAG_VERSION)" ./build_and_push_all.sh

elb-controller:
	cd files/elb-controller/container && \
	VERSION="$(AWS_TAG_VERSION)" \
	DOCKER_REPO="docker4x/l4controller-aws" \
	DOCKER_TAG="$(AWS_TAG_VERSION)" \
	DOCKER_TAG_LATEST=true \
	make -k container

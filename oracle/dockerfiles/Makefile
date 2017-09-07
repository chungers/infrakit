.PHONY: lb-controller

default: lb-controller
	VERSION="$(ORACLE_TAG_VERSION)" ./build_and_push_all.sh

elb-controller:
	@echo "Building LB Controller"
	cd lb-controller/files/container && \
	VERSION="$(ORACLE_TAG_VERSION)" \
	DOCKER_REPO="docker4x/l4controller-aws" \
	DOCKER_TAG="$(ORACLE_TAG_VERSION)" \
	DOCKER_TAG_LATEST=true \
	make -k container

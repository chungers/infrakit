.PHONY: olb-controller

default: olb-controller
	VERSION="$(ORACLE_TAG_VERSION)" ./build_and_push_all.sh

olb-controller:
	@echo "Building OLB Controller"
	cd olb-controller/files/container && \
	VERSION="$(ORACLE_TAG_VERSION)" \
	DOCKER_REPO="docker4x/l4controller-oracle" \
	DOCKER_TAG="$(ORACLE_TAG_VERSION)" \
	DOCKER_TAG_LATEST=true \
	make -k container

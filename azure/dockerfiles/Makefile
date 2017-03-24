default: alb-controller
	VERSION="$(AZURE_TAG_VERSION)" ./build_and_push_all.sh

agent:
	make -C walinuxagent clean
	make -C walinuxagent push TAG="$(AZURE_TAG_VERSION)"

alb-controller:
	cd files/alb-controller/container && \
	VERSION="$(AZURE_TAG_VERSION)" \
	DOCKER_REPO="docker4x/l4controller-azure" \
	DOCKER_TAG="$(AZURE_TAG_VERSION)" \
	DOCKER_TAG_LATEST=true \
	make -k bin/loadbalancer


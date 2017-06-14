# Makefile for building the loadbalancer controller

# Build version metadata
VERSION?=$(EDITIONS_VERSION)
REVISION=$(shell git rev-list -1 HEAD)

NAMESPACE?=docker4x
DOCKER_REPO?=${NAMESPACE}/l4controller-azure
DOCKER_TAG?=dev
DOCKER_IMAGE?=${DOCKER_REPO}:${DOCKER_TAG}

.PHONY: clean bin/loadbalancer all
.DEFAULT: all
all: bin/loadbalancer

clean:
	@echo "+ $@"
	rm -rf bin
	@# Remove local images
	docker rmi -f ${DOCKER_IMAGE} >/dev/null 2>&1 || true
	docker rmi -f ${DOCKER_REPO}:latest >/dev/null 2>&1 || true
	docker rmi -f alb-controller:build >/dev/null 2>&1 || true

# This one runs outside
bin/loadbalancer:
	echo "Building $(VERSION), Image=$(DOCKER_IMAGE)"
	cd ../../../../../ && docker build --pull -t alb-controller:build -f ${CURDIR}/Dockerfile.build .
	docker run --name alb-controller-build alb-controller:build
	mkdir -p ${CURDIR}/bin
	docker cp alb-controller-build:/go/bin/loadbalancer ${CURDIR}/bin
	-docker rm alb-controller-build
	docker build --pull -t $(DOCKER_IMAGE) -t $(DOCKER_REPO):latest .
ifeq (${DOCKER_PUSH},true)
	docker push ${DOCKER_IMAGE}
ifeq (${DOCKER_TAG_LATEST},true)
	docker push ${DOCKER_REPO}:latest
endif
endif

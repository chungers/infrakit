# Makefile for building the loadbalancer controller

# Build version metadata
VERSION?=azure-v$(EDITIONS_VERSION)
REVISION=$(shell git rev-list -1 HEAD)

NAMESPACE?=docker4x
DOCKER_REPO?=${NAMESPACE}/l4controller-azure
DOCKER_TAG?=dev
DOCKER_IMAGE?=${DOCKER_REPO}:${DOCKER_TAG}

.PHONY: clean all
.DEFAULT: all
all: govendor bin/loadbalancer

govendor:
	@echo "+ getting govendor"
	-go get -u github.com/kardianos/govendor
	govendor sync

clean:
	@echo "+ $@"
	rm -rf bin
	@# Remove local images
	docker rmi -f ${DOCKER_IMAGE} >/dev/null 2>&1 || true
	docker rmi -f ${DOCKER_REPO}:latest >/dev/null 2>&1 || true

# This target runs inside the container
build-in-container: govendor
	@echo "+ $@"
	mkdir -p /go/build
	cd ../ && \
	go build -ldflags "-X main.Version=$(VERSION) -X main.Revision=$(REVISION)" \
		-o /go/build/loadbalancer cmd/*.go

# This one runs outside
bin/loadbalancer:
	echo "Building $(VERSION), Image=$(DOCKER_IMAGE)"
	-docker rm alb-controller-build
	cd ../../../../../ && docker build --pull -t alb-controller:build -f ${CURDIR}/Dockerfile.build .
	docker run --name alb-controller-build alb-controller:build
	mkdir -p ${CURDIR}/bin
	docker cp alb-controller-build:/go/build/loadbalancer ${CURDIR}/bin
	docker build --pull -t $(DOCKER_IMAGE) -t $(DOCKER_REPO):latest -t $(DOCKER_REPO):$(VERSION) .
ifeq (${DOCKER_PUSH},1)
	docker push ${DOCKER_IMAGE}
ifeq (${DOCKER_TAG_LATEST},true)
	docker push ${DOCKER_REPO}:latest
endif
endif

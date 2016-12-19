# Makefile for building the loadbalancer controller

# Build version metadata
VERSION?=gcp-v$(EDITIONS_VERSION)
REVISION=$(shell git rev-list -1 HEAD)

DOCKER_REPO_OWNER?=docker4x
DOCKER_REPO?=${DOCKER_REPO_OWNER}/l4controller-gcp
DOCKER_TAG?=dev
DOCKER_IMAGE?=${DOCKER_REPO}:${DOCKER_TAG}

.PHONY: clean all container
.DEFAULT: all
all: govendor container

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
container:
	echo "Building $(VERSION), Image=$(DOCKER_IMAGE)"
	-docker rm lb-controller-build
	cd ../../../../../ && docker build --no-cache --pull -t lb-controller:build -f ${CURDIR}/Dockerfile.build .
	docker run \
		-e EDITIONS_VERSION=$(EDITIONS_VERSION) \
		-e VERSION=$(VERSION) \
		-e DOCKER_REPO_OWNER=$(DOCKER_REPO_OWNER) \
		-e DOCKER_REPO=$(DOCKER_REPO) \
		-e DOCKER_TAG=$(DOCKER_TAG) \
		-e DOCKER_IMAGE=$(DOCKER_IMAGE) \
		--name lb-controller-build lb-controller:build
	mkdir -p ${CURDIR}/bin
	docker cp lb-controller-build:/go/build/loadbalancer ${CURDIR}/bin
	docker build --no-cache --pull -t $(DOCKER_IMAGE) -t $(DOCKER_REPO):latest -t $(DOCKER_REPO):$(VERSION) .
ifeq (${DOCKER_PUSH},true)
	docker push ${DOCKER_IMAGE}
ifeq (${DOCKER_TAG_LATEST},true)
	docker push ${DOCKER_REPO}:latest
endif
endif

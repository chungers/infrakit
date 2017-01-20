DOCKER_REPO_OWNER ?= docker4x
DOCKER_REPO ?= ${DOCKER_REPO_OWNER}/l4controller-gcp
DOCKER_TAG ?= latest
DOCKER_IMAGE ?= ${DOCKER_REPO}:${DOCKER_TAG}
DOCKER_BUILD_FLAGS ?= --no-cache --pull
DOCKER_PUSH ?= true

.PHONY: clean all container
.DEFAULT: all
all: govendor container

govendor:
	@echo "+ getting govendor"
	-go get -u github.com/kardianos/govendor
	govendor sync

clean:
	rm -rf bin

# This target runs inside the container
build-in-container: govendor
	@echo "+ $@"
	mkdir -p /go/build
	go build -ldflags "-X main.Version=$(VERSION) -X main.Revision=$(REVISION)" \
		-o /go/build/loadbalancer cmd/*.go

# This one runs outside
container:
	echo "Building $(VERSION), Image=$(DOCKER_IMAGE)"
	-docker rm lb-controller-build
	cd ../../.. && docker build $(DOCKER_BUILD_FLAGS) -t lb-controller:build -f ${CURDIR}/Dockerfile.build .
	docker run \
		-e DOCKER_REPO_OWNER=$(DOCKER_REPO_OWNER) \
		-e DOCKER_REPO=$(DOCKER_REPO) \
		-e DOCKER_TAG=$(DOCKER_TAG) \
		-e DOCKER_IMAGE=$(DOCKER_IMAGE) \
		--name lb-controller-build lb-controller:build
	mkdir -p ${CURDIR}/bin
	docker cp lb-controller-build:/go/build/loadbalancer ${CURDIR}/bin
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_IMAGE) .
ifeq ($(DOCKER_PUSH),true)
	docker push $(DOCKER_IMAGE)
endif

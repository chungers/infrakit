DOCKER_REPO_OWNER ?= docker4x
DOCKER_REPO ?= ${DOCKER_REPO_OWNER}/lb-controller-gcp
EDITIONS_VERSION ?= latest
DOCKER_IMAGE ?= ${DOCKER_REPO}:${EDITIONS_VERSION}

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
	-docker container rm lb-controller-build
	cd ../../.. && docker build -t lb-controller:build -f ${CURDIR}/Dockerfile.build .
	docker container run \
		-e DOCKER_REPO_OWNER=$(DOCKER_REPO_OWNER) \
		-e DOCKER_REPO=$(DOCKER_REPO) \
		-e DOCKER_TAG=$(EDITIONS_VERSION) \
		-e DOCKER_IMAGE=$(DOCKER_IMAGE) \
		--name lb-controller-build lb-controller:build
	mkdir -p ${CURDIR}/bin
	docker container cp lb-controller-build:/go/build/loadbalancer ${CURDIR}/bin
	docker build -t $(DOCKER_IMAGE) .

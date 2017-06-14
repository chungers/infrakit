NAMESPACE ?= docker4x
DOCKER_REPO ?= ${NAMESPACE}/lb-controller-gcp
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
	docker rmi -f ${DOCKER_IMAGE} >/dev/null 2>&1 || true
	docker rmi -f ${DOCKER_REPO}:latest >/dev/null 2>&1 || true
	-docker rm lb-controller-build

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
	docker container run --name lb-controller-build lb-controller:build
	mkdir -p ${CURDIR}/bin
	docker container cp lb-controller-build:/go/build/loadbalancer ${CURDIR}/bin
	-docker rm -fv lb-controller-build
	docker build -t $(DOCKER_IMAGE) -t $(DOCKER_REPO):latest .
ifeq (${DOCKER_PUSH},true)
	docker push $(DOCKER_IMAGE)
ifeq ($(DOCKER_TAG_LATEST),true)
	docker push $(DOCKER_REPO):latest
endif
endif

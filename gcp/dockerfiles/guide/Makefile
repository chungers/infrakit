DOCKER_REPO_OWNER ?= docker4x
DOCKER_REPO ?= ${DOCKER_REPO_OWNER}/guide-gcp
DOCKER_TAG ?= latest
DOCKER_IMAGE ?= ${DOCKER_REPO}:${DOCKER_TAG}
DOCKER_BUILD_FLAGS ?= --no-cache --pull
DOCKER_PUSH ?= true

.PHONY: all build push test
.DEFAULT: all
all: build test push

build:
	@docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_IMAGE) .

test:
	@docker run --rm \
		-v ${CURDIR}/tests:/tests \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /usr/bin/docker:/usr/bin/docker \
		$(DOCKER_IMAGE) /tests/run.sh

push:
ifeq ($(DOCKER_PUSH),true)
	@docker push $(DOCKER_IMAGE)
endif

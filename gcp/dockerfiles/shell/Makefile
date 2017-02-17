DOCKER_REPO_OWNER ?= docker4x
DOCKER_REPO ?= ${DOCKER_REPO_OWNER}/shell-gcp
DOCKER_TAG ?= latest
DOCKER_IMAGE ?= ${DOCKER_REPO}:${DOCKER_TAG}
DOCKER_PUSH ?= true

.PHONY: all build push
.DEFAULT: all
all: build push

build:
	@docker build -t $(DOCKER_IMAGE) .

push:
ifeq ($(DOCKER_PUSH),true)
	@docker push $(DOCKER_IMAGE)
endif

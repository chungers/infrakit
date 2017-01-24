DOCKER_REPO_OWNER ?= docker4x
DOCKER_REPO ?= ${DOCKER_REPO_OWNER}/infrakit-gcp
DOCKER_TAG ?= latest
DOCKER_IMAGE ?= ${DOCKER_REPO}:${DOCKER_TAG}
DOCKER_BUILD_FLAGS ?= --no-cache --pull
DOCKER_PUSH ?= true

.PHONY: all build push
.DEFAULT: all
all: build push

build:
	@./extract_binaries.sh
	@docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_IMAGE) .

push:
ifeq ($(DOCKER_PUSH),true)
	@docker push $(DOCKER_IMAGE)
endif

DOCKER_IMAGE ?= docker4x/shell-gcp

.PHONY: all build
.DEFAULT: all
all: build

build:
	@docker build -t $(DOCKER_IMAGE) .

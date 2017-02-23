DOCKER_IMAGE ?= docker4x/infrakit-gcp

.PHONY: all build
.DEFAULT: all
all: build

build:
	@./extract_binaries.sh
	@docker build -t $(DOCKER_IMAGE) .

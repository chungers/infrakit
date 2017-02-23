DOCKER_IMAGE ?= docker4x/guide-gcp

.PHONY: all build test
.DEFAULT: all
all: build test

build:
	@docker build -t $(DOCKER_IMAGE) .

test:
	@docker container run --rm \
		-v ${CURDIR}/tests:/tests \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /usr/bin/docker:/usr/bin/docker \
		$(DOCKER_IMAGE) /tests/run.sh


DOCKER_IMAGE := docker4x/mobyhead:latest

default: build

build:
	docker build --build-arg USERNAME=$(GITHUB_USER) --build-arg TOKEN=$(GITHUB_TOKEN) -t $(DOCKER_IMAGE) .

push:
	docker push $(DOCKER_IMAGE)

release: clean build push

clean:
	docker image rm $(DOCKER_IMAGE)

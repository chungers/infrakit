NAMESPACE ?= docker4x
IMAGE := $(NAMESPACE)/rtf
DOCKER := /usr/bin/docker
DOCKER_SOCK := /var/run/docker.sock
DOCKER_VERSION ?= 17.06.1-ce
RUN_RTF := docker container run -v $(DOCKER):$(DOCKER) -v $(DOCKER_SOCK):$(DOCKER_SOCK) -t $(IMAGE)

push: build
	docker image push $(IMAGE)

build:
	docker build --build-arg DOCKER_VERSION=$(DOCKER_VERSION) -t $(IMAGE) .

run: build
	$(RUN_RTF) -v run

run-verbose: build
	$(RUN_RTF) -vvv run

clean: clean-tmp-files clean-containers clean-images-all

clean-tmp-files:
	find . -name "*~" -type f -delete

clean-containers:
	IDS=$$(docker ps -a | grep ${IMAGE} | awk '{ print $$1 }'); for container in $$(echo $$IDS); do docker rm $$container; done

clean-images-all: clean-images-old
	IDS=$$(docker image ls | grep $(IMAGE) | awk '{ print $$3 }'); for image in $$(echo $$IDS); do docker rmi $$image; done

clean-images-old:
	IDS=$$(docker image ls | grep "<none>" | awk '{ print $$3 }'); for image in $$(echo $$IDS); do docker rmi $$image; done

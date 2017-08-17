
NAMESPACE ?= docker4x
DOCKER := /usr/bin/docker
DOCKER_SOCK := /var/run/docker.sock
RUN_RTF := docker container run -v $(DOCKER):$(DOCKER) -v $(DOCKER_SOCK):$(DOCKER_SOCK) -t $(NAMESPACE)/rtf

push: build
	docker image push $(NAMESPACE)/rtf

build:
	docker build -t $(NAMESPACE)/rtf .

run: build
	$(RUN_RTF)

run-vvv: build
	$(RUN_RTF) -vvv

clean:
	find . -name "*~" -type f -delete
DOCKER := /usr/bin/docker
DOCKER_SOCK := /var/run/docker.sock

push: build
	docker image push corbincoleman/docker_rtf

build:
	docker build -t corbincoleman/docker_rtf .

run:
	docker container run -v $(DOCKER):$(DOCKER) -v $(DOCKER_SOCK):$(DOCKER_SOCK) corbincoleman/docker_rtf
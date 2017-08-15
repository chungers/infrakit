NAMESPACE := corbincoleman
DOCKER := /usr/bin/docker
DOCKER_SOCK := /var/run/docker.sock

push: build
	docker image push $(NAMESPACE)/rtf

build:
	docker build -t $(NAMEPSACE)/rtf .

run:
	docker container run -v $(DOCKER):$(DOCKER) -v $(DOCKER_SOCK):$(DOCKER_SOCK) $(NAMESPACE)/rtf run

run-vvv:
	docker container run -v $(DOCKER):$(DOCKER) -v $(DOCKER_SOCK):$(DOCKER_SOCK) $(NAMESPACE)/rtf -vvv run

run-it:
	docker container run -v $(DOCKER):$(DOCKER) -v $(DOCKER_SOCK):$(DOCKER_SOCK) --entrypoint /bin/sh -it $(NAMESPACE)/rtf

clean:
	find . -name "*~" -type f -delete
DOCKER_IMAGE_NAME := swarm-exec
DOCKER_CONTAINER_NAME := ${DOCKER_IMAGE_NAME}-build-container

default: build

build: clean	
	docker build -t ${DOCKER_IMAGE_NAME} -f Dockerfile.build .
	docker create --name ${DOCKER_CONTAINER_NAME} ${DOCKER_IMAGE_NAME}
	mkdir -p bin
	docker cp ${DOCKER_CONTAINER_NAME}:/go/bin/swarm-exec bin/swarm-exec
	-docker rm -f ${DOCKER_CONTAINER_NAME}

release:
	./release.sh

go-build:
	go build

clean:
	-docker rm -f ${DOCKER_CONTAINER_NAME}
	-rm bin/swarm-exec

DOCKER_IMAGE_NAME :=buoy
DOCKER_CONTAINER_NAME :=buoy-build-container

default: build

build:
	-docker rm -f ${DOCKER_CONTAINER_NAME}
	docker build -t ${DOCKER_IMAGE_NAME} -f Dockerfile.build .
	docker run --name ${DOCKER_CONTAINER_NAME} ${DOCKER_IMAGE_NAME}
	mkdir -p bin
	docker cp ${DOCKER_CONTAINER_NAME}:/go/bin/buoy bin/buoy
	-docker rm -f ${DOCKER_CONTAINER_NAME}

clean:
	-rm bin/buoy

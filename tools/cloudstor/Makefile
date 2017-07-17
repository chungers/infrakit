DOCKER_IMAGE_NAME := cloudstor
DOCKER_CONTAINER_NAME := ${DOCKER_IMAGE_NAME}-build-container_$(shell date +"%s")

default: rootfs

clean:
	-rm -rf ./plugin
	-rm cloudstor-rootfs.tar.gz
	-docker rm -vf ${DOCKER_CONTAINER_NAME}
	-rm bin/cloudstor

build:
	docker build -t ${DOCKER_IMAGE_NAME} -f Dockerfile.build .
	docker create --name ${DOCKER_CONTAINER_NAME} ${DOCKER_IMAGE_NAME}
	mkdir -p bin
	docker cp ${DOCKER_CONTAINER_NAME}:/go/bin/cloudstor bin/cloudstor
	docker rm -vf ${DOCKER_CONTAINER_NAME}
	docker rmi ${DOCKER_IMAGE_NAME}

rootfs:  clean build
	docker build -t ${NAMESPACE}/cloudstor:rootfs .
	mkdir -p ./plugin/rootfs
	docker create --name ${DOCKER_CONTAINER_NAME} ${NAMESPACE}/cloudstor:rootfs
	docker export ${DOCKER_CONTAINER_NAME} | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf ${DOCKER_CONTAINER_NAME}
	tar zcf cloudstor-rootfs.tar.gz ./plugin/

ifeq (${NAMESPACE},)
	NAMESPACE := docker4x
endif
ifeq (${TAG_VERSION},)
	TAG_VERSION := latest
endif

DOCKER_IMAGE_NAME := cloudstor
DOCKER_CONTAINER_NAME := ${DOCKER_IMAGE_NAME}-build-container

default: push

clean:
	-rm -rf ./plugin
	-rm cloudstor-rootfs.tar.gz
	-docker rm -vf tmp
	-rm bin/cloudstor


build:
	docker build -q -t ${DOCKER_IMAGE_NAME} -f Dockerfile.build .
	docker create --name tmp ${DOCKER_IMAGE_NAME}
	mkdir -p bin
	docker cp tmp:/go/bin/cloudstor bin/cloudstor
	docker rm -vf tmp
	docker rmi ${DOCKER_IMAGE_NAME}

rootfs:  clean build
	docker build -q -t ${NAMESPACE}/cloudstor:rootfs .
	mkdir -p ./plugin/rootfs
	docker create --name tmp ${NAMESPACE}/cloudstor:rootfs
	docker export tmp | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf tmp
	tar zcvf cloudstor-rootfs.tar.gz ./plugin/

create:  clean build rootfs
	-docker plugin rm -f ${NAMESPACE}/cloudstor:${TAG_VERSION}
	docker plugin create ${NAMESPACE}/cloudstor:${TAG_VERSION} ./plugin

push:  clean build rootfs create
ifeq (${DOCKER_PUSH},1)
	docker plugin push ${NAMESPACE}/cloudstor:${TAG_VERSION}
endif

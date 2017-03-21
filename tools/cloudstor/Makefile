ifeq (${NAMESPACE},)
	NAMESPACE := docker4x
endif
ifeq (${TAG_VERSION},)
	TAG_VERSION := latest
endif

default: push

clean:
	-rm -rf ./plugin
	-rm cloudstor-rootfs.tar.gz
	-docker rm -vf tmp


build:
	docker build -q -t cloudstor-builder -f Dockerfile.dev .
	docker create --name tmp cloudstor-builder
	docker cp tmp:/go/bin/cloudstor .
	docker rm -vf tmp
	docker rmi cloudstor-builder

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

PLUGIN_NAME ?= docker4x/cloudstor
PLUGIN_TAG ?= latest

default: push

clean:
	rm -rf ./plugin
	docker rm -vf tmp || true


build:
	docker build -q -t cloudstor-builder -f Dockerfile.dev .
	docker create --name tmp cloudstor-builder
	docker cp tmp:/go/bin/cloudstor .
	docker rm -vf tmp
	docker rmi cloudstor-builder

rootfs:
	docker build -q -t ${PLUGIN_NAME}:rootfs .
	mkdir -p ./plugin/rootfs
	docker create --name tmp ${PLUGIN_NAME}:rootfs
	docker export tmp | tar -x -C ./plugin/rootfs
	cp config.json ./plugin/
	docker rm -vf tmp

create:
	docker plugin rm -f ${PLUGIN_NAME}:${PLUGIN_TAG} || true
	docker plugin create ${PLUGIN_NAME}:${PLUGIN_TAG} ./plugin

push:  clean build rootfs create
	docker plugin push ${PLUGIN_NAME}:${PLUGIN_TAG}

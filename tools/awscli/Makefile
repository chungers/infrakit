NAMESPACE := docker4x
VERSION := latest

default: build push

build:
	docker build -t ${NAMESPACE}/awscli:${VERSION} .

push:
ifeq (${DOCKER_PUSH},true)
	docker push ${NAMESPACE}/awscli:${VERSION}
endif



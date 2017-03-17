NAMESPACE := docker4x
VERSION := latest

default: build push

build:
	docker build -t ${NAMESPACE}/awscli:${VERSION} .

push:
	docker push ${NAMESPACE}/awscli:${VERSION}



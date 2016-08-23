SERVICE := diagnosticspipeline
IMAGE := nathanleclaire/diagnosticspipeline
BUCKETVOL := diagnosticsbucket
BUCKETTARGET := /bucket

default: build

build:
	docker build -t $(IMAGE) .

run: volume
	docker service create \
		--name $(SERVICE) \
		--mount type=bind,source=$(HOME)/.aws/credentials,target=/root/aws/.credentials \
		--mount type=bind,source=$(HOME)/.docker,target=/root/.docker \
		--mount type=bind,source=$(shell which docker),target=/bin/docker \
		--mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
		--mount type=volume,source=$(BUCKETVOL),target=$(BUCKETTARGET) \
		$(IMAGE)

volumes:
	docker volume create --name $(BUCKETVOL)

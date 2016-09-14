SERVICE := diagnosticspipeline
IMAGE := docker4x/diagnosticspipeline
BUCKETVOL := diagnosticsbucket

default: build

build:
	docker build -t $(IMAGE) .

run: build
	docker run -ti \
		--rm \
		--name $(SERVICE) \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e DEBUG=1 \
		-e LIMIT="-30 second" \
		-v $(HOME)/.docker:/root/.docker \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(shell which docker):/bin/docker  \
		$(IMAGE)

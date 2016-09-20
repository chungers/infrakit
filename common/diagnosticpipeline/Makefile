SERVICE := diagnosticspipeline
IMAGE := editions/diagnosticspipeline
BUCKETVOL := diagnosticsbucket

default: build

build:
	docker build -t $(IMAGE) .

run: build
	docker run -ti \
		--rm \
		--log-opt max-size=50m \
		--log-opt max-file=2 \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e DEBUG=1 \
		-e LIMIT="-30 second" \
		-v $(HOME)/.docker:/root/.docker \
		-v /var/run/docker.sock:/var/run/docker.sock \
		$(IMAGE)

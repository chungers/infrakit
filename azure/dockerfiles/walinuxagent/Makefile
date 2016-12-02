TAG := dev

default: build

build:
	docker build -t docker4x/agent-azure:$(TAG) .

push: build
	docker push docker4x/agent-azure:$(TAG)

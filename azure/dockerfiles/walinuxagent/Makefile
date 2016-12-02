TAG := azure-v1.13.0-rc2-beta12

default: build

build:
	docker build -t docker4x/agent-azure:$(TAG) .

push: build
	docker push docker4x/agent-azure:$(TAG)

WORK_DIR := /root/editions/azure
BIND_DIR := $(shell pwd)

login: credsdir
	docker run -ti -v azurecredentials:/root/.azure azuresdk/azure-cli-python:latest az login

shell: credsdir
	docker run -ti -w $(WORK_DIR) -v $(BIND_DIR):$(WORK_DIR) -v azurecredentials:/root/.azure azuresdk/azure-cli-python:latest

credsdir:
	docker volume create --name azurecredentials || true

clean:
	docker rmi azuresdk/azure-cli-python:latest

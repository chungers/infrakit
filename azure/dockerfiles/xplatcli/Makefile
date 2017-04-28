WORK_DIR := /root/editions/azure
BIND_DIR := $(shell pwd)
AZ_CLI := azuresdk/azure-cli-python@sha256:30e9ef97ace1b8ad2fdeadcfc52e49b317e0f81515d26957b1c9ae670b8a0aae

login: credsdir
	docker run -ti -v azurecredentials:/root/.azure $(AZ_CLI) az login

shell: credsdir
	docker run -ti -w $(WORK_DIR) -v $(BIND_DIR):$(WORK_DIR) -v azurecredentials:/root/.azure $(AZ_CLI)

credsdir:
	docker volume create --name azurecredentials || true

clean:
	docker rmi azuresdk/azure-cli-python:latest

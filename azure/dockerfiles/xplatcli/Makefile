WORK_DIR := /root/editions/azure
BIND_DIR := $(shell pwd)
AZ_CLI := azuresdk/azure-cli-python@sha256:30e9ef97ace1b8ad2fdeadcfc52e49b317e0f81515d26957b1c9ae670b8a0aae

login: credsdir
	docker run -ti -v azurecredentials:/root/.azure $(AZ_CLI) az login

shell: login
	docker run -ti -w $(WORK_DIR) -v $(BIND_DIR):$(WORK_DIR) -v azurecredentials:/root/.azure $(AZ_CLI)


testsub: login
	docker run -it -v azurecredentials:/root/.azure $(AZ_CLI) az account set --subscription d5d89bd9-4ef0-4187-a4ed-5de123bd9520

copy_blob: testsub
	docker run -ti -v azurecredentials:/root/.azure -e AZURE_STORAGE_KEY -e AZURE_STORAGE_ACCOUNT=edition $(AZ_CLI) az storage blob copy start \
		--source-blob $(BLOB_NAME) \
		--source-container mobylinux \
		--source-account-name $(STG_ACCOUNT) \
		--destination-blob $(BLOB_NAME) \
		--destination-container mobylinux

credsdir:
	docker volume create --name azurecredentials || true

clean:
	docker rmi azuresdk/azure-cli-python:latest

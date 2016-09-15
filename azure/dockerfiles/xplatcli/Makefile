WORK_DIR := /root/editions/azure
BIND_DIR := $(shell pwd)

build:
	docker build -t docker4x/cli-azure .

login: credsdir build
	docker run -ti -v azurecredentials:/root/.azure docker4x/cli-azure login

shell: credsdir build
	docker run -ti -v azurecredentials:/root/.azure docker4x/cli-azure config mode arm
	docker run -ti -w $(WORK_DIR) -v $(BIND_DIR):$(WORK_DIR) -v azurecredentials:/root/.azure --entrypoint sh docker4x/cli-azure

credsdir:
	docker volume create --name azurecredentials || true

clean:
	docker rmi docker4x/cli-azure

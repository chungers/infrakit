# Copyright 2016 Docker Inc. All rights reserved.

DEPLOYMENT_NAME ?= docker
GCLOUD_IMAGE_TAG = sha256:44296dea85d3f8f016a9ac2d3225ac494c8cfc4593b0d3fb316eb26413e5846d
GCLOUD = docker run --rm -ti \
	-v $(CURDIR)/templates:/templates \
	--volumes-from gcloud-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gcloud
DEPLOYMENTS = $(GCLOUD) deployment-manager deployments

.PHONY: all auth revoke create describe delete
.DEFAULT: all
all: auth create

auth: revoke
	docker run -it --name gcloud-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) true
	$(GCLOUD) init --skip-diagnostics

revoke:
	docker rm gcloud-config 2>/dev/null || true

create:
	$(DEPLOYMENTS) create $(DEPLOYMENT_NAME) --config /templates/swarm.jinja

describe:
	$(DEPLOYMENTS) describe docker

delete:
	$(GCLOUD) compute instances delete $(shell $(GCLOUD) compute instances list --filter='tags.items ~ $(DEPLOYMENT_NAME)-node' --uri)
	$(DEPLOYMENTS) delete $(DEPLOYMENT_NAME)

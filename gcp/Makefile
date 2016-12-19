# Copyright 2016 Docker Inc. All rights reserved.

GCLOUD_IMAGE_TAG = sha256:44296dea85d3f8f016a9ac2d3225ac494c8cfc4593b0d3fb316eb26413e5846d
GCLOUD = docker run --rm -ti \
	-v $(CURDIR)/templates:/templates \
	--volumes-from gcloud-config \
	google/cloud-sdk@$(GCLOUD_IMAGE_TAG) \
	gcloud
DEPLOYMENTS = $(GCLOUD) deployment-manager deployments

all: auth create

auth: revoke
	docker run -it --name gcloud-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) true
	$(GCLOUD) init --skip-diagnostics

revoke:
	docker rm gcloud-config 2>/dev/null || true

create:
	$(DEPLOYMENTS) create docker --config /templates/swarm.py

describe:
	$(DEPLOYMENTS) describe docker

delete:
	$(GCLOUD) compute instances delete $$(gcloud compute instances list --filter='tags.items ~ swarm' --uri)
	$(DEPLOYMENTS) delete docker

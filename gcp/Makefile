# Copyright 2016 Docker Inc. All rights reserved.

GCLOUD_IMAGE_TAG = sha256:3a4e572df0c716196b3ffdeacaba3ebdd36f8cdfdac6d5ee4bd50cd405366736
GCLOUD = docker run --rm -ti -v $(CURDIR)/configuration:/configuration --volumes-from gcloud-config google/cloud-sdk gcloud
DEPLOYMENTS = $(GCLOUD) deployment-manager deployments

all: auth create

auth: revoke
	docker run -it --name gcloud-config google/cloud-sdk@$(GCLOUD_IMAGE_TAG) true
	$(GCLOUD) init

revoke:
	docker rm gcloud-config || true

create:
	$(DEPLOYMENTS) create docker --config configuration/docker.yaml

describe:
	$(DEPLOYMENTS) describe docker

delete:
	$(DEPLOYMENTS) delete docker -q

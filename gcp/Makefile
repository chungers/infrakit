# Copyright 2016 Docker Inc. All rights reserved.

auth:
	docker run -it --name gcloud-config google/cloud-sdk gcloud init

revoke:
	docker rm -f gcloud-config

create:
	docker run --rm -ti -v $(CURDIR)/configuration:/configuration --volumes-from gcloud-config google/cloud-sdk gcloud deployment-manager deployments create docker --config configuration/docker.yaml

describe:
	docker run --rm -ti -v $(CURDIR)/configuration:/configuration --volumes-from gcloud-config google/cloud-sdk gcloud deployment-manager deployments describe docker

delete:
	docker run --rm -ti -v $(CURDIR)/configuration:/configuration --volumes-from gcloud-config google/cloud-sdk gcloud deployment-manager deployments delete docker -q

.PHONY: release clean build template

default: release

build:
	mkdir -p tmp/ || true
	mkdir -p ../../dist/ || true
	docker build -t aws-release .

template: build
	docker run -ti --rm -u $(shell id -u) -v $(PWD)/dist:/home/docker/dist/ aws-release --docker_version="$(EDITIONS_DOCKER_VERSION)" --edition_tag="$(EDITIONS_TAG)" --ami_id="$(AWS_AMI_LIST)" --ami_src_region="$(AMI_SRC_REGION)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL_CLOUD)" --account_list_url="$(DOCKER_AWS_ACCOUNT_URL)" --public="$(MAKE_AMI_PUBLIC)" --share="$(SHARE_AMI)"

release: clean build
	docker run -ti --rm -u $(shell id -u) -v $(PWD)/dist:/home/docker/dist/ -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY aws-release --upload --docker_version="$(DOCKER_VERSION)" --edition_tag="$(EDITION_TAG)" --ami_id="$(AWS_AMI_LIST)" --ami_src_region="$(AMI_SRC_REGION)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL_CLOUD)" --account_list_url="$(DOCKER_AWS_ACCOUNT_URL)" --public="$(MAKE_AMI_PUBLIC)" --share="$(SHARE_AMI)"

clean:
	docker rmi aws-release || true

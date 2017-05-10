.PHONY: release clean build template

default: release

build:
	-mkdir -p tmp/ $(ROOT_DIR)/dist/
	docker build -t aws-release -f cfn_gen/Dockerfile cfn_gen

template: build
	docker run -ti --rm -u $(shell id -u) -v $(ROOT_DIR)/dist:/home/docker/dist/ -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY aws-release --docker_version="$(DOCKER_VERSION)" --edition_tag="$(EDITIONS_TAG)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL)-cloud"

clean:
	docker rmi aws-release || true

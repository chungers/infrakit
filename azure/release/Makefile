.PHONY: release clean build template

default: release

build:
	mkdir -p tmp/ || true 
	mkdir -p ../../dist/ || true
	cp ../templates/editions.template.json tmp/editions.template
	docker build -t azure-release .

template: clean build
	docker run -ti --rm -v $(PWD)/dist:/home/docker/dist/ azure-release --docker_version="$(EDITIONS_DOCKER_VERSION)" --edition_version="$(EDITIONS_VERSION)" --vhd_sku="$(VHD_SKU)" --vhd_version="$(VHD_VERSION)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL_CLOUD)"

release: clean build
	docker run -ti --rm -v $(PWD)/dist:/home/docker/dist/ azure-release --upload --docker_version="$(EDITIONS_DOCKER_VERSION)" --edition_version="$(EDITIONS_VERSION)" --vhd_sku="$(VHD_SKU)" --vhd_version="$(VHD_VERSION)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL_CLOUD)"

clean:
	# rm -f tmp/*
	# rm -Rf dist/*
	docker rmi azure-release || true

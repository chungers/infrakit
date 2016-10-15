DOCKER_VERSION := $(DOCKER_VERSION)
EDITION_VERSION := $(EDITION_VERSION)
VHD_SKU := $(VHD_SKU)
VHD_VERSION := $(VHD_VERSION)
CHANNEL := $(CHANNEL)

build:
	docker build -t azure-release .

release: clean build
	docker run -ti --rm -v $(PWD)/release/dist:/home/docker/dist/ azure-release --docker_version="$(DOCKER_VERSION)" --edition_version="$(EDITION_VERSION)" --vhd_sku="$(VHD_SKU)" --vhd_version="$(VHD_VERSION)" --channel="$(CHANNEL)"

clean:
	docker rmi azure-release && ([ $$? -eq 0 ] && echo "success!") || echo "failure!"
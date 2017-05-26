.PHONY: release clean build template

default: build

build: clean
	mkdir -p tmp/ || true
	-mkdir -p ${ROOT_DIR}/dist
	cp ../templates/editions.template.json tmp/editions.template
	docker build -t azure-release .

template: build
	docker run --rm -u $(shell id -u) -v ${ROOT_DIR}/dist:/home/docker/dist/ azure-release --docker_version="$(DOCKER_VERSION)" --editions_version="$(EDITIONS_VERSION)" --vhd_sku="$(VHD_SKU)" --vhd_version="$(VHD_VERSION)" --offer_id="$(VHD_OFFER_ID)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL_CLOUD)" --channel_ddc="$(CHANNEL_DDC)" --ee_vhd_sku="$(EE_VHD_SKU)" --ee_vhd_version="$(EE_VHD_VERSION)" --ee_offer_id="$(EE_OFFER_ID)" --edition_addon=$(EDITION_ADDON)

clean:
	# rm -f tmp/*
	# rm -Rf dist/*
	-docker rmi azure-release

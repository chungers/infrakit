.PHONY: release clean build template

default: release

build:
	mkdir -p tmp/ || true
	mkdir -p ../../dist/ || true
	cp ../templates/editions.template.json tmp/editions.template
	docker build -t azure-release .

template: build
	docker run -ti --rm -u $(shell id -u) -v $(PWD)/dist:/home/docker/dist/ azure-release --docker_version="$(EDITIONS_DOCKER_VERSION)" --edition_version="$(EDITIONS_VERSION)" --vhd_sku="$(VHD_SKU)" --vhd_version="$(VHD_VERSION)" --offer_id="$(VHD_OFFER_ID)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL_CLOUD)" --channel_ddc="$(CHANNEL_DDC)" --ee_vhd_sku="$(EE_VHD_SKU)" --ee_vhd_version="$(EE_VHD_VERSION)" --ee_offer_id="$(EE_OFFER_ID)" --edition_addon=$(EDITION_ADDON)

release: clean build
	docker run -ti --rm -u $(shell id -u) -v $(PWD)/dist:/home/docker/dist/ -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY azure-release --upload --docker_version="$(EDITIONS_DOCKER_VERSION)" --edition_version="$(EDITIONS_VERSION)" --vhd_sku="$(VHD_SKU)" --vhd_version="$(VHD_VERSION)" --offer_id="$(VHD_OFFER_ID)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL_CLOUD)" --channel_ddc="$(CHANNEL_DDC)" --ee_vhd_sku="$(EE_VHD_SKU)" --ee_vhd_version="$(EE_VHD_VERSION)"  --ee_offer_id="$(EE_OFFER_ID)" --edition_addon=$(EDITION_ADDON)

clean:
	# rm -f tmp/*
	# rm -Rf dist/*
	docker rmi azure-release || true

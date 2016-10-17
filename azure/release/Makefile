.PHONY: release clean build

default: release

build:
	mkdir -p ../../tmp/ || true 
	mkdir -p ../../dist/ || true
	cp ../templates/editions.template.json tmp/editions.template
	docker build -t azure-release .

release: clean build
	docker run -ti --rm -e AWS_ACCESS_KEY_ID="AKIAJA5ETQ6INL6FNP5Q" -e AWS_SECRET_ACCESS_KEY="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" -v $(PWD)/dist:/home/docker/dist/ azure-release --docker_version="$(EDITIONS_DOCKER_VERSION)" --edition_version="$(EDITIONS_VERSION)" --vhd_sku="$(VHD_SKU)" --vhd_version="$(VHD_VERSION)" --channel="$(CHANNEL)" --channel_cloud="$(CHANNEL_CLOUD)"

clean:
	# rm -f tmp/*
	# rm -Rf dist/*
	docker rmi azure-release || true

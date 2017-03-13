#!/bin/bash
set -e

BASE_DIR=../../

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
VERSION="${VERSION:-latest}"

for IMAGE in init guide create-sp ddc-init cloud logger meta
do
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-azure:${VERSION}"
	docker build --pull -t "${FINAL_IMAGE}" -f "Dockerfile.${IMAGE}" .
	docker push "${FINAL_IMAGE}"
done

# Ensure that this image (meant to be interacted with manually) has :latest tag
# as well as a specific one
docker tag "${NAMESPACE}/create-sp-azure:${VERSION}" "${NAMESPACE}/create-sp-azure:latest"
docker push "${NAMESPACE}/create-sp-azure:latest"

pushd $BASE_DIR/common/cloudstor
PLUGIN_TAG=${VERSION} make
popd

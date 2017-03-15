#!/bin/bash
set -e

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

# build and push cloudstor plugin
tar zxvf cloudstor-rootfs.tar.gz
docker plugin rm -f "${NAMESPACE}/cloudstor:${VERSION}" || true
docker plugin create "${NAMESPACE}/cloudstor:${VERSION}" ./plugin
docker plugin push "${NAMESPACE}/cloudstor:${VERSION}"

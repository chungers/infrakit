#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
TAG_VERSION="${AZURE_TAG_VERSION:-latest}"

for IMAGE in init guide create-sp ddc-init cloud logger meta
do
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-azure:${TAG_VERSION}"
	docker build --pull -t "${FINAL_IMAGE}" -f "Dockerfile.${IMAGE}" .
	if [ ${DOCKER_PUSH} -eq 1 ]; then
		docker push "${FINAL_IMAGE}"
	fi
done

# Ensure that this image (meant to be interacted with manually) has :latest tag
# as well as a specific one
docker tag "${NAMESPACE}/create-sp-azure:${TAG_VERSION}" "${NAMESPACE}/create-sp-azure:latest"
if [ ${DOCKER_PUSH}-eq 1 ]; then
	docker push "${NAMESPACE}/create-sp-azure:latest"
fi


# build and push cloudstor plugin
tar zxvf cloudstor-rootfs.tar.gz -C files/
docker plugin rm -f "${NAMESPACE}/cloudstor:${TAG_VERSION}" || true
docker plugin create "${NAMESPACE}/cloudstor:${TAG_VERSION}" ./plugin
if [ ${DOCKER_PUSH} -eq 1 ]; then
	docker plugin push "${NAMESPACE}/cloudstor:${TAG_VERSION}"
fi

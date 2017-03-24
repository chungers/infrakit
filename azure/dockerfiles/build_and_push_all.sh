#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
TAG_VERSION="${AZURE_TAG_VERSION:-latest}"

CURR_DIR=`pwd`
ROOTDIR="${ROOTDIR:-$CURR_DIR}"
DEFAULT_PATH="dist/azure/nightly/$TAG_VERSION"
AZURE_TARGET_PATH="${AZURE_TARGET_PATH:-$DEFAULT_PATH}"

echo "+ Creating dist folder: $AZURE_TARGET_PATH"
mkdir -p $ROOTDIR/$AZURE_TARGET_PATH

for IMAGE in init guide create-sp ddc-init cloud logger meta
do
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-azure:${TAG_VERSION}"
	docker build --pull -t "${FINAL_IMAGE}" -f "Dockerfile.${IMAGE}" .
	if [ ${IMAGE} != "ddc-init" ] && [ "${IMAGE}" != "cloud" ]; then
		echo "++ Saving docker image to: ${ROOTDIR}/${AZURE_TARGET_PATH}/${IMAGE}-azure.tar"
		docker save "${FINAL_IMAGE}" > "${ROOTDIR}/${AZURE_TARGET_PATH}/${IMAGE}-azure.tar"
	fi
	if [ ${DOCKER_PUSH} -eq 1 ]; then
		docker push "${FINAL_IMAGE}"
	fi
done

# build and push walinuxagent image
pushd walinuxagent
docker build --pull -t docker4x/agent-azure:${TAG_VERSION} .
echo "++ Saving docker image to: ${ROOTDIR}/${AZURE_TARGET_PATH}/agent-azure.tar"
docker save "docker4x/agent-azure:${TAG_VERSION}" > "${ROOTDIR}/${AZURE_TARGET_PATH}/agent-azure.tar"
if [ ${DOCKER_PUSH} -eq 1 ]; then
	docker push "docker4x/agent-azure:${TAG_VERSION}"
fi
popd


# Ensure that this image (meant to be interacted with manually) has :latest tag
# as well as a specific one
docker tag "${NAMESPACE}/create-sp-azure:${TAG_VERSION}" "${NAMESPACE}/create-sp-azure:latest"
if [ "${DOCKER_PUSH}" -eq 1 ]; then
	docker push "${NAMESPACE}/create-sp-azure:latest"
fi


# build and push cloudstor plugin
pushd files
tar zxvf cloudstor-rootfs.tar.gz
docker plugin rm -f "${NAMESPACE}/cloudstor:${TAG_VERSION}" || true
docker plugin create "${NAMESPACE}/cloudstor:${TAG_VERSION}" ./plugin
rm -rf ./plugin
if [ ${DOCKER_PUSH} -eq 1 ]; then
	docker plugin push "${NAMESPACE}/cloudstor:${TAG_VERSION}"
fi
popd

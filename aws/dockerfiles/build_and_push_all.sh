#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
TAG_VERSION="${AWS_TAG_VERSION:-latest}"
CURR_DIR=`pwd`
ROOTDIR="${ROOTDIR:-$CURR_DIR}"
DEFAULT_PATH="dist/aws/nightly/$TAG_VERSION"
AWS_TARGET_PATH="${AWS_TARGET_PATH:-$DEFAULT_PATH}"

echo "+ Creating dist folder: $AWS_TARGET_PATH"
mkdir -p $ROOTDIR/$AWS_TARGET_PATH

for IMAGE in shell init guide ddc-init cloud meta
do
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-aws:${TAG_VERSION}"
	docker build --pull -t "${FINAL_IMAGE}" -f "Dockerfile.${IMAGE}" .
	echo "++ Saving docker image to: ${ROOTDIR}/${AWS_TARGET_PATH}/${IMAGE}-aws.tar"
	docker save "${FINAL_IMAGE}" > "${ROOTDIR}/${AWS_TARGET_PATH}/${IMAGE}-aws.tar"
	if [ "${DOCKER_PUSH}" -eq 1 ]; then
		docker push "${FINAL_IMAGE}"
	fi
done

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
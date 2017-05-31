#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
TAG_VERSION="${AZURE_TAG_VERSION:-latest}"

## expects EDITIONS_DOCKER_VERSION in YY.MM.X format
UPGRADE_TAG="${EDITIONS_DOCKER_VERSION:0:5}-latest"

CURR_DIR=`pwd`
ROOT_DIR="${ROOT_DIR:-$CURR_DIR}"
DEFAULT_PATH="dist/azure/nightly/$TAG_VERSION"
AZURE_TARGET_PATH="${AZURE_TARGET_PATH:-$DEFAULT_PATH}"

echo -e "+ \033[1mCreating dist folder:\033[0m $AZURE_TARGET_PATH"
mkdir -p $ROOT_DIR/$AZURE_TARGET_PATH


# Test all images built
function test () {
	if [ -z "$1" ]
	then
     echo "Image to test is needed"
		 exit 1
	fi
	IMAGE=${1}
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-azure:${TAG_VERSION}"
	echo -e "+++ \033[1mTesting\033[0m \033[4m${FINAL_IMAGE}\033[0m"
	docker container run --rm \
		-v ${CURR_DIR}/${IMAGE}/tests:/tests \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /usr/bin/docker:/usr/bin/docker \
		--entrypoint sh \
		${FINAL_IMAGE} /tests/run.sh
}

#copy in common files that apply across containers
for IMAGE in init guide ddc-init logger
do
	cp common/* ${IMAGE}/files/
done

for IMAGE in init guide create-sp ddc-init cloud logger meta
do
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-azure:${TAG_VERSION}"
	echo -e "++ \033[1mBuilding image:\033[0m ${FINAL_IMAGE}"
	docker build --pull -t "${FINAL_IMAGE}" -f "${IMAGE}/Dockerfile" ${IMAGE}
	if [ ${IMAGE} != "ddc-init" ] && [ "${IMAGE}" != "cloud" ]; then
		echo -e "++ \033[1mSaving docker image to:\033[0m ${ROOT_DIR}/${AZURE_TARGET_PATH}/${IMAGE}-azure.tar"
		docker save "${FINAL_IMAGE}" > "${ROOT_DIR}/${AZURE_TARGET_PATH}/${IMAGE}-azure.tar"
	fi
	test ${IMAGE}
	if [ ${DOCKER_PUSH} -eq 1 ]; then
		docker push "${FINAL_IMAGE}"
	fi
done

# build and push walinuxagent image
docker build --pull -t docker4x/agent-azure:${TAG_VERSION} -f walinuxagent/Dockerfile walinuxagent
echo -e "++ \033[1mSaving docker image to:\033[0m ${ROOT_DIR}/${AZURE_TARGET_PATH}/agent-azure.tar"
docker save "docker4x/agent-azure:${TAG_VERSION}" > "${ROOT_DIR}/${AZURE_TARGET_PATH}/agent-azure.tar"
if [ ${DOCKER_PUSH} -eq 1 ]; then
	docker push "docker4x/agent-azure:${TAG_VERSION}"
fi


# Ensure that this image (meant to be interacted with manually) has :latest tag
# as well as a specific one
docker tag "${NAMESPACE}/create-sp-azure:${TAG_VERSION}" "${NAMESPACE}/create-sp-azure:latest"
if [ "${DOCKER_PUSH}" -eq 1 ]; then
	docker push "${NAMESPACE}/create-sp-azure:latest"
fi

# Build upgrade-azure passing in the necessary env vars
docker build --pull -t docker4x/upgrade-azure:${TAG_VERSION} --build-arg VERSION=${EDITIONS_DOCKER_VERSION} --build-arg CHANNEL=${CHANNEL} -f upgrade/Dockerfile upgrade
# Ensure that the upgrade image has :YY.MM-latest tag as well so that 
# upgrade.sh in shell can easily refer to it
docker tag "${NAMESPACE}/upgrade-azure:${TAG_VERSION}" "${NAMESPACE}/upgrade-azure:${UPGRADE_TAG}"
if [ "${DOCKER_PUSH}" -eq 1 ]; then
	docker push "${NAMESPACE}/upgrade-azure:${UPGRADE_TAG}"
fi

# build and push cloudstor plugin
tar zxf cloudstor-rootfs.tar.gz
docker plugin rm -f "${NAMESPACE}/cloudstor:${TAG_VERSION}" || true
docker plugin create "${NAMESPACE}/cloudstor:${TAG_VERSION}" ./plugin
rm -rf ./plugin
if [ ${DOCKER_PUSH} -eq 1 ]; then
	docker plugin push "${NAMESPACE}/cloudstor:${TAG_VERSION}"
fi

#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
TAG_VERSION="${AWS_TAG_VERSION:-latest}"

CURR_DIR=`pwd`
ROOT_DIR="${ROOT_DIR:-$CURR_DIR}"
DEFAULT_PATH="dist/aws/nightly/$TAG_VERSION"
AWS_TARGET_PATH="${AWS_TARGET_PATH:-$DEFAULT_PATH}"

# Test all images built
function test () {
	if [ -z "$1" ]
	then
     echo "Image to test is needed"
		 exit 1
	fi
	IMAGE=${1}
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-aws:${TAG_VERSION}"
	echo -e "+++ \033[1mTesting\033[0m \033[4m${FINAL_IMAGE}\033[0m"
	docker container run --rm \
		-v ${CURR_DIR}/${IMAGE}/tests:/tests \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /usr/bin/docker:/usr/bin/docker \
		--entrypoint sh \
		${FINAL_IMAGE} /tests/run.sh
}

echo -e "+ \033[1mCreating dist folder:\033[0m $AWS_TARGET_PATH"
# Create directory and make sure to chmod it
mkdir -p $ROOT_DIR/$AWS_TARGET_PATH 
if [ $? ]; then
	docker run --rm -v $ROOT_DIR:/data alpine sh -c "chmod +rwx -R /data/dist"
	mkdir -p $ROOT_DIR/$AWS_TARGET_PATH 
fi

for IMAGE in shell init guide ddc-init cloud meta
do
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-aws:${TAG_VERSION}"
	echo -e "++ \033[1mBuilding image:\033[0m ${FINAL_IMAGE}"
	docker build --pull -t "${FINAL_IMAGE}" -f "${IMAGE}/Dockerfile" ${IMAGE}
	if [ ${IMAGE} != "ddc-init" ] && [ "${IMAGE}" != "cloud" ]; then
		echo -e "++ \033[1mSaving docker image to:\033[0m ${ROOT_DIR}/${AWS_TARGET_PATH}/${IMAGE}-aws.tar"
		docker save "${FINAL_IMAGE}" --output "${ROOT_DIR}/${AWS_TARGET_PATH}/${IMAGE}-aws.tar"
	fi
	test ${IMAGE}
	if [ "${DOCKER_PUSH}" = true ]; then
		docker push "${FINAL_IMAGE}"
	fi
done

# build and push cloudstor plugin
tar zxf cloudstor-rootfs.tar.gz
docker plugin rm -f "${NAMESPACE}/cloudstor:${TAG_VERSION}" || true
docker plugin create "${NAMESPACE}/cloudstor:${TAG_VERSION}" ./plugin
rm -rf ./plugin
if [ "${DOCKER_PUSH}" = true ]; then
	docker plugin push "${NAMESPACE}/cloudstor:${TAG_VERSION}"
fi

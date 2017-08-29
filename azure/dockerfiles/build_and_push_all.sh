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
# Create directory and make sure to chmod it
mkdir -p $ROOT_DIR/$AZURE_TARGET_PATH 
if [ $? ]; then
  docker run --rm -v $ROOT_DIR:/data alpine sh -c "chmod +rwx -R /data/dist"
  mkdir -p $ROOT_DIR/$AZURE_TARGET_PATH 
fi


# Test all images built
function check_image () {
  if [ -z "$1" ]
  then
     echo "Image to test is needed"
     exit 1
  fi
  TAG=${1#*:}
  NAMESPACE_IMAGE=${1%:*}
  IMAGE=${NAMESPACE_IMAGE#*/}
  FOLDER=${IMAGE%*-azure}
  echo -e "+++ \033[1mTesting\033[0m \033[4m${1}\033[0m"
  docker container run --rm \
    -v ${CURR_DIR}/${FOLDER}/tests:/tests \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /usr/bin/docker:/usr/bin/docker \
    --entrypoint sh \
    ${FINAL_IMAGE} /tests/run.sh
}

function docker_tag_exists() {
  if [ -z "$1" ]
  then
     echo "Image to test is needed"
     exit 1
  fi
  TAG=${1#*:}
  NAMESPACE_IMAGE=${1%:*}
  IMAGE=${NAMESPACE_IMAGE#*/}
  FOLDER=${IMAGE%*-azure}
  EXISTS=$(curl -f -slSL https://hub.docker.com/v2/repositories/${NAMESPACE_IMAGE}/tags/?page_size=10000 | jq -r "[.results | .[] | .name == \"${TAG}\"] | any")
  #test $EXISTS = true
  true
}

#copy in common files that apply across containers
for IMAGE in init guide ddc-init logger upgrade lookup
do
  cp common/* ${IMAGE}/files/
done

for IMAGE in init guide create-sp ddc-init cloud logger meta lookup
do
  FINAL_IMAGE="${NAMESPACE}/${IMAGE}-azure:${TAG_VERSION}"
  echo -e "++ \033[1mBuilding image:\033[0m ${FINAL_IMAGE}"
  docker build --pull -t "${FINAL_IMAGE}" -f "${IMAGE}/Dockerfile" ${IMAGE}
  check_image ${FINAL_IMAGE}
  if [ "${DOCKER_PUSH}" = true ]; then
    docker push "${FINAL_IMAGE}"
    if ! docker_tag_exists "${FINAL_IMAGE}"; then
      echo -e "+++ \033[31mError - Final Image tag not found! ${FINAL_IMAGE}\033[0m"
      exit 1
    fi
  fi
done

# build and push walinuxagent image
docker build --pull -t docker4x/agent-azure:${TAG_VERSION} -f walinuxagent/Dockerfile walinuxagent
if [ "${DOCKER_PUSH}" = true ]; then
  docker push "docker4x/agent-azure:${TAG_VERSION}"
  if ! docker_tag_exists "${FINAL_IMAGE}"; then
    echo -e "+++ \033[31mError - Final Image tag not found! ${FINAL_IMAGE}\033[0m"
    exit 1
  fi
fi


# Ensure that this image (meant to be interacted with manually) has :latest tag
# as well as a specific one
docker tag "${NAMESPACE}/create-sp-azure:${TAG_VERSION}" "${NAMESPACE}/create-sp-azure:latest"
if [ "${DOCKER_PUSH}" -eq 1 ]; then
  docker push "${NAMESPACE}/create-sp-azure:latest"
fi

# Build upgrade-azure-core passing in the necessary env vars
docker build --pull -t docker4x/upgrade-core-azure:${TAG_VERSION} --build-arg VERSION=${EDITIONS_DOCKER_VERSION} --build-arg CHANNEL=${CHANNEL} -f upgrade/Dockerfile upgrade
# Build upgrade-azure wrapper that will invoke upgrade-azure-core without users having to be aware of the Customdata mount
docker build --pull -t docker4x/upgrade-azure:${UPGRADE_TAG} --build-arg TAG_VERSION=${TAG_VERSION} -f upgrade/Dockerfile.wrapper upgrade

# Ensure that the upgrade image has :YY.MM-latest tag as well so that upgrade.sh in shell can easily refer to it
docker tag "${NAMESPACE}/upgrade-azure-core:${TAG_VERSION}" "${NAMESPACE}/upgrade-azure-core:${UPGRADE_TAG}"
if [ "${DOCKER_PUSH}" = true ]; then
  # this is for advanced users/internal folks who may want to test upgrade to every Editions release
  docker push "${NAMESPACE}/upgrade-azure-core:${TAG_VERSION}"
  # this is for quick/easy reference to YY.MM.X-latest using the upgrade.sh wrapper
  docker push "${NAMESPACE}/upgrade-azure-core:${UPGRADE_TAG}"
  # this is mainly for 17.03/early 17.06 EE users who do not have upgrade.sh in the shell
  docker push "${NAMESPACE}/upgrade-azure:${UPGRADE_TAG}"
fi

# build and push cloudstor plugin
tar zxf cloudstor-rootfs.tar.gz
docker plugin rm -f "${NAMESPACE}/cloudstor:${TAG_VERSION}" || true
docker plugin create "${NAMESPACE}/cloudstor:${TAG_VERSION}" ./plugin
rm -rf ./plugin
if [ "${DOCKER_PUSH}" = true ]; then
  docker plugin push "${NAMESPACE}/cloudstor:${TAG_VERSION}"
fi

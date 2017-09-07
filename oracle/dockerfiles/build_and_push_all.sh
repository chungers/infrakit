#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
TAG_VERSION="${ORACLE_TAG_VERSION:-latest}"
COMPOSE_VERSION="${COMPOSE_VERSION:-1.15.0}"

CURR_DIR=`pwd`
ROOT_DIR="${ROOT_DIR:-$CURR_DIR}"
DEFAULT_PATH="dist/oracle/nightly/$TAG_VERSION"
ORACLE_TARGET_PATH="${ORACLE_TARGET_PATH:-$DEFAULT_PATH}"

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
  FOLDER=${IMAGE%*-oracle} 
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
  FOLDER=${IMAGE%*-oracle}
  EXISTS=$(curl -f -slSL https://hub.docker.com/v2/repositories/${NAMESPACE_IMAGE}/tags/?page_size=10000 | jq -r "[.results | .[] | .name == \"${TAG}\"] | any")
  #test $EXISTS = true
  true
}

echo -e "+ \033[1mCreating dist folder:\033[0m $ORACLE_TARGET_PATH"
# Create directory and make sure to chmod it
mkdir -p $ROOT_DIR/$ORACLE_TARGET_PATH 
if [ $? ]; then
  docker run --rm -v $ROOT_DIR:/data alpine sh -c "chmod +rwx -R /data/dist"
  mkdir -p $ROOT_DIR/$ORACLE_TARGET_PATH 
fi

for IMAGE in shell init guide ddc-init cloud meta
do
  FINAL_IMAGE="${NAMESPACE}/${IMAGE}-oracle:${TAG_VERSION}"
  echo -e "++ \033[1mBuilding image:\033[0m ${FINAL_IMAGE}"
  BUILD_ARGS=""
  if [ "$IMAGE" = "shell" ]; then
    BUILD_ARGS="--build-arg COMPOSE_VERSION=$COMPOSE_VERSION"
  fi
  docker build --pull -t "${FINAL_IMAGE}" $ARGS -f "${IMAGE}/Dockerfile" ${IMAGE}
  check_image ${FINAL_IMAGE}
  if [ "${DOCKER_PUSH}" = true ]; then
    docker push "${FINAL_IMAGE}"
    if ! docker_tag_exists "${FINAL_IMAGE}"; then
      echo -e "+++ \033[31mError - Final Image tag not found! ${FINAL_IMAGE}\033[0m"
      exit 1
    fi
  fi
done

# # build and push cloudstor plugin
# tar zxf cloudstor-rootfs.tar.gz
# docker plugin rm -f "${NAMESPACE}/cloudstor:${TAG_VERSION}" || true
# docker plugin create "${NAMESPACE}/cloudstor:${TAG_VERSION}" ./plugin
# rm -rf ./plugin
# if [ "${DOCKER_PUSH}" = true ]; then
#   docker plugin push "${NAMESPACE}/cloudstor:${TAG_VERSION}"
# fi

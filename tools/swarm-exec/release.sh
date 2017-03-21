#!/bin/bash
set -x

TAG_VERSION=${TAG_VERSION:-"v0.1"}
NAMESPACE=${NAMESPACE:-"docker4x"}
IMAGE_NAME=${IMAGE_NAME:-"swarm-exec"}
DOCKER_TAG_LATEST=${DOCKER_TAG_LATEST:-"yes"}
FULL_IMAGE_NAME=$NAMESPACE/$IMAGE_NAME

# build the binary
./build.sh

docker build -t $FULL_IMAGE_NAME:$TAG_VERSION -f Dockerfile .
if [ ${DOCKER_PUSH} -eq 1 ]; then
    docker push $FULL_IMAGE_NAME:$TAG_VERSION

    if [[ "$DOCKER_TAG_LATEST" == "yes" ]] ; then
        docker push $FULL_IMAGE_NAME:latest
    fi
fi

rm -f bin/swarm-exec

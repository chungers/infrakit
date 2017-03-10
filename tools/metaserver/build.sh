#!/bin/bash

set -e

DOCKER_IMAGE_NAME="metaserver"
DOCKER_CONTAINER_NAME="metaserver-build-container"

if [[ $(docker ps -a | grep $DOCKER_CONTAINER_NAME) != "" ]]; then
    echo "remove $DOCKER_CONTAINER_NAME"
    docker rm -f $DOCKER_CONTAINER_NAME
fi

docker build -t $DOCKER_IMAGE_NAME .

docker run --name $DOCKER_CONTAINER_NAME $DOCKER_IMAGE_NAME ./compile.sh

mkdir -p bin
docker cp $DOCKER_CONTAINER_NAME:/go/bin/metaserver bin/metaserver

if [[ $(docker ps -a | grep $DOCKER_CONTAINER_NAME) != "" ]]; then
    echo "remove $DOCKER_CONTAINER_NAME"
    docker rm -f $DOCKER_CONTAINER_NAME
fi

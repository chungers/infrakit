#!/bin/bash

set -e

DOCKER_IMAGE_NAME="buoy"
DOCKER_CONTAINER_NAME="buoy-build-container"

if [[ $(docker ps -a | grep $DOCKER_CONTAINER_NAME) != "" ]]; then
  docker rm -f $DOCKER_CONTAINER_NAME
fi

docker build -t $DOCKER_IMAGE_NAME -f Dockerfile.build .

docker run --name $DOCKER_CONTAINER_NAME $DOCKER_IMAGE_NAME

mkdir -p bin
docker cp $DOCKER_CONTAINER_NAME:/go/bin/buoy bin/buoy

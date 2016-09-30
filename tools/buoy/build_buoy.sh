#!/bin/bash

set -e

DOCKER_IMAGE_NAME="buoy"
DOCKER_CONTAINER_NAME="buoy-build-container"

if [[ $(docker ps -a | grep $DOCKER_CONTAINER_NAME) != "" ]]; then
  docker rm -f $DOCKER_CONTAINER_NAME
fi

docker build -t $DOCKER_IMAGE_NAME -f Dockerfile.buoy .

docker run --name $DOCKER_CONTAINER_NAME $DOCKER_IMAGE_NAME ./compile_buoy.sh


mkdir -p bin
docker cp $DOCKER_CONTAINER_NAME:/go/bin/buoy bin/buoy

mkdir -p ../../aws/dockerfiles/files/bin/ ../../azure/dockerfiles/files/bin/
cp bin/buoy ../../aws/dockerfiles/files/bin/buoy
cp bin/buoy ../../azure/dockerfiles/files/bin/buoy

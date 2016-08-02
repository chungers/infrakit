#!/bin/bash

set -e

DOCKER_IMAGE_NAME="aws-buoy"
DOCKER_CONTAINER_NAME="aws-buoy-build-container"
TAG="RC3"

if [[ $(docker ps -a | grep $DOCKER_CONTAINER_NAME) != "" ]]; then
  docker rm -f $DOCKER_CONTAINER_NAME 2>/dev/null
fi

docker build -t $DOCKER_IMAGE_NAME:$TAG -f Dockerfile.buoy .

docker run --name $DOCKER_CONTAINER_NAME $DOCKER_IMAGE_NAME:$TAG ./compile_buoy.sh


mkdir -p bin
docker cp $DOCKER_CONTAINER_NAME:/go/bin/buoy bin/buoy

mkdir -p ../../aws/dockerfiles/files/bin/
cp bin/buoy ../../aws/dockerfiles/files/bin/buoy

mkdir -p ../../azure/dockerfiles/files/bin/
cp bin/buoy ../../azure/dockerfiles/files/bin/buoy

#!/bin/bash

set -e 

rm -rf build src packages/aws/dockerimages/
mkdir -p build/aws build/azure build/gcp src packages/aws/dockerimages/

docker rm moby || true
docker create --name moby mobylinux/media:$MOBY_IMG_COMMIT ls
docker cp moby:/initrd.img src/initrd.img
docker cp moby:/vmlinuz64 build/
docker rm moby

if [ $LOAD_IMAGES == "true" ]; then
	echo "+ Copying Docker images: cp $ROOTDIR/$AWS_TARGET_PATH/*.tar packages/aws/dockerimages/"
	cp $ROOTDIR/$AWS_TARGET_PATH/*.tar packages/aws/dockerimages/
fi

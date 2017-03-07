#!/bin/bash

set -e

mkdir -p build/aws build/azure build/gcp src packages/aws/dockerimages/

docker rm moby || true
docker create --name moby mobylinux/media:${MOBY_IMG_COMMIT} true
docker cp moby:/initrd.img src/initrd.img
docker cp moby:/vmlinuz64 build/
docker rm moby

# Bundle the docker images if we're deploying to Marketplace
if [ "$LOAD_IMAGES" == "true" ]; then
	echo "+ Copying Docker images: from $ROOTDIR/$AWS_TARGET_PATH/*.tar to packages/aws/dockerimages/"		
	cp $ROOTDIR/$AWS_TARGET_PATH/*.tar packages/aws/dockerimages/		
	echo "+ Remove DDC and Cloud images from dockerimages"
	rm -f packages/aws/dockerimages/ddc-init-aws.tar packages/aws/dockerimages/cloud-aws.tar || true
fi

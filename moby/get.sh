#!/bin/bash

set -e

mkdir -p build/aws build/azure build/gcp src packages/aws/var/dockerimages/ packages/azure/var/dockerimages/ packages/gcp/var/dockerimages/ tmp

echo "++ Copying Moby build ${MOBY_IMG_URL} to ${MOBY_IMG_NAME}"
docker run --rm -v ${PWD}/tmp:/tmp docker4x/awscli:latest s3 --no-sign-request cp ${MOBY_IMG_URL} /tmp/${MOBY_IMG_NAME} 
pushd tmp
tar xvf ${MOBY_IMG_NAME}
cp initrd.img ../src/initrd.img
cp vmlinuz64 ../build/
popd
rm -rf tmp

# Bundle the docker images if we're deploying to Marketplace
if [ "$LOAD_IMAGES" == "true" ]; then
	if [ -e "$ROOTDIR/$AWS_TARGET_PATH/shell-aws.tar" ]; then
		echo "++ Copying Docker images: from $ROOTDIR/$AWS_TARGET_PATH/*.tar to packages/aws/var/dockerimages/"
		cp $ROOTDIR/$AWS_TARGET_PATH/*.tar packages/aws/var/dockerimages/
	fi
	if [ -e "$ROOTDIR/$AZURE_TARGET_PATH/agent-azure.tar" ]; then
		echo "++ Copying Docker images: from $ROOTDIR/$AZURE_TARGET_PATH/*.tar to packages/azure/var/dockerimages/"
		cp $ROOTDIR/$AZURE_TARGET_PATH/*.tar packages/azure/var/dockerimages/
	fi
else
	if [ -e "$ROOTDIR/$AWS_TARGET_PATH/shell-aws.tar" ]; then
		echo "++ Copying Docker Shell image: from $ROOTDIR/$AWS_TARGET_PATH/shell-aws.tar to packages/aws/var/dockerimages/"
		cp $ROOTDIR/$AWS_TARGET_PATH/shell-aws.tar packages/aws/var/dockerimages/
	fi
	if [ -e "$ROOTDIR/$AZURE_TARGET_PATH/agent-azure.tar" ]; then
		echo "++ Copying Docker Agent image: from $ROOTDIR/$AZURE_TARGET_PATH/agent-azure.tar to packages/azure/var/dockerimages/"
		cp $ROOTDIR/$AZURE_TARGET_PATH/agent-azure.tar packages/azure/var/dockerimages/
	fi
fi

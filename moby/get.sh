#!/bin/bash

set -e

mkdir -p build/aws build/azure build/gcp src packages/aws/var/dockerimages/ tmp

echo "++ Copying Moby build ${MOBY_IMG_URL} to ${MOBY_IMG_NAME}"
docker run --rm -v ${PWD}/tmp:/tmp -e AWS_ACCESS_KEY_ID=$QA_AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$QA_AWS_SECRET_ACCESS_KEY docker4x/awscli:latest s3 cp ${MOBY_IMG_URL} /tmp/${MOBY_IMG_NAME} 
pushd tmp
tar xvf ${MOBY_IMG_NAME}
cp initrd.img ../src/initrd.img
cp vmlinuz64 ../build/
popd
rm -rf tmp

# Bundle the docker images if we're deploying to Marketplace
if [ "$LOAD_IMAGES" == "true" ]; then
	echo "++ Copying Docker images: from $ROOTDIR/$AWS_TARGET_PATH/*.tar to packages/aws/var/dockerimages/"		
	cp $ROOTDIR/$AWS_TARGET_PATH/*.tar packages/aws/var/dockerimages/
	echo "++ Copying Docker images: from $ROOTDIR/$AZURE_TARGET_PATH/*.tar to packages/azure/var/dockerimages/"		
	cp $ROOTDIR/$AZURE_TARGET_PATH/*.tar packages/azure/var/dockerimages/
else
	echo "++ Copying Docker Shell image: from $ROOTDIR/$AWS_TARGET_PATH/shell-aws.tar to packages/aws/var/dockerimages/"		
	cp $ROOTDIR/$AWS_TARGET_PATH/shell-aws.tar packages/aws/var/dockerimages/
	echo "++ Copying Docker Agent image: from $ROOTDIR/$AZURE_TARGET_PATH/agent-azure.tar to packages/azure/var/dockerimages/"		
	cp $ROOTDIR/$AZURE_TARGET_PATH/agent-azure.tar packages/azure/var/dockerimages/
fi

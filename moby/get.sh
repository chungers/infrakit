#!/bin/bash

set -e

mkdir -p build/aws build/azure build/gcp src packages/aws/dockerimages/ packages/azure/dockerimages/ packages/gcp/dockerimages/ tmp

echo "++ Copying Moby build ${MOBY_IMG_URL} to ${MOBY_IMG_NAME}"
docker run --rm -v ${PWD}/tmp:/tmp docker4x/awscli:latest s3 --quiet --no-sign-request cp ${MOBY_IMG_URL} /tmp/${MOBY_IMG_NAME}
pushd tmp
tar xvf ${MOBY_IMG_NAME}
cp initrd.img ../src/initrd.img
cp vmlinuz64 ../build/
popd
rm -rf tmp

# Bundle ALL docker images if we're deploying to Marketplace
# Only include the shell containers everything else can be pulled down
echo "++ Check if shell exists at $ROOT_DIR/$AWS_TARGET_PATH/shell-aws.tar"
if [ -e "$ROOT_DIR/$AWS_TARGET_PATH/shell-aws.tar" ]; then
	echo "++ Copying Docker Shell image: from $ROOT_DIR/$AWS_TARGET_PATH/shell-aws.tar to packages/aws/dockerimages/"
	cp $ROOT_DIR/$AWS_TARGET_PATH/shell-aws.tar packages/aws/dockerimages/
else
	echo "++ MISSING Docker Shell image: $ROOT_DIR/$AWS_TARGET_PATH/shell-aws.tar"
	ls -ltr $ROOT_DIR/dist/*
	exit 1
fi
echo "++ Check if agent exists at $ROOT_DIR/$AZURE_TARGET_PATH/agent-azure.tar"
if [ -e "$ROOT_DIR/$AZURE_TARGET_PATH/agent-azure.tar" ]; then
	echo "++ Copying Docker Agent image: from $ROOT_DIR/$AZURE_TARGET_PATH/agent-azure.tar to packages/azure/dockerimages/"
	cp $ROOT_DIR/$AZURE_TARGET_PATH/agent-azure.tar packages/azure/dockerimages/
else
	echo "++ MISSING Docker Agent image: $ROOT_DIR/$AZURE_TARGET_PATH/agent-azure.tar"
	ls -ltr $ROOT_DIR/dist/*
	exit 1
fi

echo "++ Check if GCP images exists at $ROOT_DIR/$GCP_TARGET_PATH/images.tar"
if [ -e ../gcp/build/images.tar ]; then
	cp ../gcp/build/images.tar packages/gcp/dockerimages/
fi
if [ -e "$ROOT_DIR/$GCP_TARGET_PATH/images.tar" ]; then
	echo "++ Copying GCP Docker image: from $ROOT_DIR/$GCP_TARGET_PATH/images.tar to packages/gcp/dockerimages/"
	cp $ROOT_DIR/$GCP_TARGET_PATH/images.tar packages/gcp/dockerimages/
fi

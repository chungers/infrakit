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

# Only include the shell containers everything else can be pulled down
echo "++ Download AWS shell and include it as shell-aws.tar"
docker pull ${AWS_SHELL} 2>&1 || echo $?
docker save ${AWS_SHELL} --output "tmp/shell-aws.tar"
echo "++ Check if shell exists at tmp/shell-aws.tar"
if [ -e "tmp/shell-aws.tar" ]; then
	echo "++ Copying Docker Shell image: from tmp/shell-aws.tar to packages/aws/dockerimages/"
	cp tmp/shell-aws.tar packages/aws/dockerimages/
else
	echo "++ MISSING Docker Shell image: tmp/shell-aws.tar"
	exit 1
fi

echo "++ Download Azure shell and include it as agent-azure.tar"
docker pull ${AZURE_SHELL} 2>&1 || echo $?
docker save ${AZURE_SHELL} --output "tmp/agent-azure.tar"
echo "++ Check if agent exists at tmp/agent-azure.tar"
if [ -e "tmp/agent-azure.tar" ]; then
	echo "++ Copying Docker Agent image: from tmp/agent-azure.tar to packages/azure/dockerimages/"
	cp tmp/agent-azure.tar packages/azure/dockerimages/
else
	echo "++ MISSING Docker Agent image: tmp/agent-azure.tar"
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

rm -rf tmp
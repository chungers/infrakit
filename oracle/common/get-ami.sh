#!/bin/bash
set -e

# get docker version
echo "Getting ${AMI_S3_PATH}/ami_id.out from $AMI_S3_BUCKET"

EDITIONS_META=$(docker run --rm -it docker4x/awscli:latest s3api --no-sign-request get-object --bucket $AMI_S3_BUCKET --key ${AMI_S3_PATH}/ami_id.out docker.out | jq -r '.Metadata')
TMP_EDITIONS_VERSION=$EDITIONS_VERSION
export EDITIONS_VERSION=$(echo $EDITIONS_META | jq -r '.editions_version')
export DOCKER_VERSION=$(echo $EDITIONS_META | jq -r '.docker_version')
export MOBY_COMMIT=$(echo $EDITIONS_META | jq -r '.moby_commit')
export AWS_TARGET_PATH=dist/aws/${CHANNEL}/${AWS_TAG_VERSION}

echo -e "+++ \033[1m${CHANNEL}\033[0m build of: \033[4m${EDITIONS_VERSION}\033[0m"

if [ -z ${EDITIONS_VERSION} ]; then
	echo "+++ No EDITIONS_VERSION found"
	exit 1
fi

echo -e "+++ Creating dist folder: $AWS_TARGET_PATH"
CURR_DIR=`pwd`
ROOT_DIR="${ROOT_DIR:-$CURR_DIR}"
AMI_OUT_DIR=$ROOT_DIR/$AWS_TARGET_PATH
mkdir -p $AMI_OUT_DIR
# Download the ami_id.out
docker run --rm -it -v $AMI_OUT_DIR:/tmp/ docker4x/awscli:latest s3 --no-sign-request cp s3://${AMI_S3_BUCKET}/${AMI_S3_PATH}/ami_id.out /tmp/$AMI_OUT
echo -e "+++ AMI build captured, lets move onto next part."

# get ami-id
AMI_ID=$(cat $AMI_OUT_DIR/ami_id.out)
if [[ -z $AMI_ID ]]
then
    echo "+++ There is no AMI_ID, nothing to do, so stopping."
    # there are no AMI's ready skip.
     exit 1
fi

echo -e "+++ \033[1mAMI\033[0m: $AMI_ID is now availble in $AMI_SOURCE_REGION"

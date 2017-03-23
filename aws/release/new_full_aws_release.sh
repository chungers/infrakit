#!/bin/bash

set -e
# This script is expected to be running on an AWS instance with a correct IAM profile to build AMI's
# also it needs AWS creds, and a docker account that is logged in that has push permissions on docker4x domain.
# run by calling ./full_aws_release.sh -d docker_version -e edition_version -b docker_binary_url -c beta -l account_list_url
export PYTHONUNBUFFERED=1
# prehook

AMI_EDITIONS_VERSION=
DOCKER_VERSION=
EDITION_VERSION=
CHANNEL=
DOCKER_AWS_ACCOUNT_URL=
MAKE_AMI_PUBLIC="no"

usage()
{
cat << EOF

usage: $0 options

This script will help release a new version of Docker for AWS.

Required ENV variables:
    - AWS_ACCESS_KEY_ID
    - AWS_SECRET_ACCESS_KEY

OPTIONS:
   -h      Show this message
   -a      AMI Edition version (17.03.0-ce, etc.)
   -b      AWS Build number (1, 2, 3, etc.)
   -c      Release Channel (beta, nightly, etc.)
   -l      AWS account list URL
   -p      Make AMI public (yes, no)
EOF
}

while getopts "hc:l:d:e:b:p:m:" OPTION
do
     case $OPTION in
         h)
             usage
             exit 1
             ;;
         a)
             AMI_EDITIONS_VERSION=$OPTARG
             ;;
         b)
             BUILD_NUMBER=$OPTARG
             ;;
         c)
             CHANNEL=$OPTARG
             ;;
         l)
             DOCKER_AWS_ACCOUNT_URL=$OPTARG
             ;;
         p)
             MAKE_AMI_PUBLIC=$OPTARG
             ;;
         ?)
             usage
             exit
             ;;
     esac
done

if [[ -z $AMI_EDITIONS_VERSION ]]
then
     usage
     exit 1
fi

if [[ -z $AWS_ACCESS_KEY_ID ]] || [[ -z $AWS_SECRET_ACCESS_KEY ]]
then
     echo ""
     echo "ERROR: The following environment variables are required, and they were not found."
     echo "    AWS_ACCESS_KEY_ID"
     echo "    AWS_SECRET_ACCESS_KEY"
     echo " "
     echo "Please add them, and try again."
     echo ""
     exit 1
fi


AMI_URL="data/ami/${AMI_EDITIONS_VERSION}/ami_id.out"

EDITIONS_META=$(aws s3api get-object --bucket $AMI_BUCKET --key ${AMI_URL} docker.out | jq -r '.Metadata')
export EDITIONS_VERSION=$(echo $EDITIONS_META | jq -r '.editions_version')
export DOCKER_VERSION=$(echo $EDITIONS_META | jq -r '.docker_version')
export AWS_TAG_VERSION=$EDITIONS_VERSION-aws${BUILD_NUMBER}

export ROOTDIR=`pwd`
export AMI_SRC_REGION=$(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region)
export HUB_LOGIN_ID=$(docker info | grep Username)
export AWS_TARGET_PATH="dist/aws/$CHANNEL/$EDITIONS_VERSION"
export RELEASE=1

echo "------"
echo "Getting started with the release...."
echo "== Parameters =="
echo "BUILD_NUMBER=$BUILD_NUMBER"
echo "DOCKER_VERSION=$DOCKER_VERSION"
echo "EDITION_VERSION=$EDITION_VERSION"
echo "EDITIONS_VERSION=$EDITIONS_VERSION"
echo "AWS_TAG_VERSION=$AWS_TAG_VERSION"
echo "AMI_EDITIONS_VERSION=$AMI_EDITIONS_VERSION"
echo "AMI_SRC_REGION=$AMI_SRC_REGION"
echo "CHANNEL=$CHANNEL"
echo "DOCKER_AWS_ACCOUNT_URL=$DOCKER_AWS_ACCOUNT_URL"
echo "Docker Hub Login ID=$HUB_LOGIN_ID"
echo "MAKE_AMI_PUBLIC=$MAKE_AMI_PUBLIC"
echo "-------"
echo "== Checking Hub login =="

# check if they have push access to docker4x org, we do this by trying to pull the private test image.
export PUSH_CHECK=$(docker pull docker4x/test 2>&1 | grep "Error")

if [[ -z $HUB_LOGIN_ID ]] || [[ -n $PUSH_CHECK ]]
then
     echo ""
     echo "ERROR: It doesn't look like you are logged into the hub, or you don't have the correct permissions"
     echo "    to push to the docker4x org"
     echo "    docker login using a docker ID that has permissions to push to the docker4x org."
     echo ""
     exit 1
fi

echo "== Checking DOCKER_BIN_URL to make sure it is valid. =="
if curl --output /dev/null --silent --head --fail "$DOCKER_BIN_URL"; then
  echo "URL exists: $DOCKER_BIN_URL"
else
  echo "URL does not exist: $DOCKER_BIN_URL, please enter a valid one."
  exit 1
fi

echo "== Checking DOCKER_AWS_ACCOUNT_URL to make sure it is valid. =="
if curl --output /dev/null --silent --head --fail "$DOCKER_AWS_ACCOUNT_URL"; then
  echo "URL exists: $DOCKER_AWS_ACCOUNT_URL"
else
  echo "URL does not exist: $DOCKER_AWS_ACCOUNT_URL, please enter a valid one."
  exit 1
fi


echo "== Checking directories =="
echo "This script assumes directories are setup in the following way"
echo "two directories at the same level. make sure correct versions are checked out (master, etc)"
echo " / "
echo " /editions  <-- github.com/docker/editions "
echo "Make sure you run this script from the editions/aws/release directory."

BASE_DIR=`pwd`

echo "-------"
echo "== Get AMI =="

# Download the ami_id.out
aws s3 cp s3://${AMI_BUCKET}/${AMI_URL} ./cloud/aws/ami_id.out
echo "AMI build captured, lets move onto next part."

CI_AMI_ID=$(cat ./cloud/aws/ami_id.out)
echo "CI_AMI_ID=$CI_AMI_ID"

# Copy AMI to desired region:
AMI_ID=$(aws ec2 copy-image --source-image-id $CI_AMI_ID --source-region us-west-2 --region $AMI_SRC_REGION --name "Moby Linux ${AWS_TAG_VERSION} ${CHANNEL}" | jq -r '.ImageId')

echo "Copied AMI to $AMI_SRC_REGION waiting for it to be available"
# Wait for AMI copy to be done
while :
do
    AMI_STATE=$(aws ec2 describe-images --image-ids $AMI_ID | jq -r '.Images[0].State')
    if [ "$AMI_STATE" == "available" ]; then
        break
    elif [ "$AMI_STATE" == "failed" ]; then
        echo "ERROR in AMI copy"
        exit 1
    fi
    echo -ne "#"
done

echo "AMI: $AMI_ID is now availble in $AMI_SRC_REGION"

# move out of the way, so it doesn't cause problems later.
mv -f ./cloud/aws/ami_id.out ./cloud/aws/ami_id.out.done

# run release, this will create CFN templates and push them to s3, push AMI to different regions and share with list of approved accounts.
./new_run_release.sh -d $DOCKER_VERSION -e $AWS_TAG_VERSION -a $AMI_ID -r $AMI_SRC_REGION -c $CHANNEL -l $DOCKER_AWS_ACCOUNT_URL -u cloud-$CHANNEL -p $MAKE_AMI_PUBLIC

echo "===== Done ====="

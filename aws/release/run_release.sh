#!/bin/bash

set -e
# run by calling ./run_release.sh docker_version edition_version ami_id ami_src_region

# prehook

BUILD_NUMBER=${BUILD_NUMBER:-24}
DOCKER_VERSION=
EDITION_VERSION=
AMI_ID=
AMI_SRC_REGION=
CHANNEL=
AWS_ACCOUNT_LIST_URL=

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
   -d      Docker version (1.12.0, etc)
   -e      Edition version (beta4, etc)
   -a      AMI ID (ami-123456, etc)
   -r      AMI source region (us-east-1)
   -c      Release Channel (beta, nightly, etc)
   -l      AWS account list URL
EOF
}

while getopts "hc:l:d:e:a:r:" OPTION
do
     case $OPTION in
         h)
             usage
             exit 1
             ;;
         d)
             DOCKER_VERSION=$OPTARG
             ;;
         e)
             EDITION_VERSION=$OPTARG
             ;;
         a)
             AMI_ID=$OPTARG
             ;;
         r)
             AMI_SRC_REGION=$OPTARG
             ;;
         c)
             CHANNEL=$OPTARG
             ;;
         l)
             AWS_ACCOUNT_LIST_URL=$OPTARG
             ;;
         ?)
             usage
             exit
             ;;
     esac
done

if [[ -z $DOCKER_VERSION ]] || [[ -z $EDITION_VERSION ]] || [[ -z $AMI_ID ]] || [[ -z $AMI_SRC_REGION ]]
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

echo "------"
echo "Getting started with the release...."
echo "== Parameters =="
echo "BUILD_NUMBER=$BUILD_NUMBER"
echo "DOCKER_VERSION=$DOCKER_VERSION"
echo "EDITION_VERSION=$EDITION_VERSION"
echo "AMI_ID=$AMI_ID"
echo "AMI_SRC_REGION=$AMI_SRC_REGION"
echo "CHANNEL=$CHANNEL"
echo "AWS_ACCOUNT_LIST_URL=$AWS_ACCOUNT_LIST_URL"
echo "-------"
echo "== Prepare files =="

# prepare the files.
mkdir -p tmp
if [ -f tmp/docker_for_aws.template ]; then
    echo "Cleanup old template file."
    rm -f tmp/docker_for_aws.template
fi
echo "Copy over template file."
cp ../cloudformation/docker_for_aws.json tmp/docker_for_aws.template

echo "== build docker image =="
# build the docker image
docker build -t docker4x/release-$BUILD_NUMBER -f Dockerfile.aws_release .

echo "== run the docker image =="
# run the image
docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
-e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
-e DOCKER_VERSION=$DOCKER_VERSION \
-e EDITION_VERSION=$EDITION_VERSION \
-e AMI_ID=$AMI_ID \
-e AMI_SRC_REGION=$AMI_SRC_REGION \
-e CHANNEL="$CHANNEL" \
-e AWS_ACCOUNT_LIST_URL="$AWS_ACCOUNT_LIST_URL" \
docker4x/release-$BUILD_NUMBER

# posthook

echo "== cleanup the docker image =="
# cleanup the image
docker rmi -f docker4x/release-$BUILD_NUMBER

echo "== finished =="
echo "------"

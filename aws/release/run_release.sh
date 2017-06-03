#!/bin/bash

set -e
# run by calling ./run_release.sh docker_version EDITIONS_VERSION ami_id ami_src_region
export PYTHONUNBUFFERED=1
# prehook

BUILD_NUMBER=${BUILD_NUMBER:-24}
DOCKER_VERSION=
EDITIONS_VERSION=
AMI_ID=
AMI_SRC_REGION=
CHANNEL=
CHANNEL_CLOUD=
AWS_ACCOUNT_LIST_URL=
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
   -d      Docker version (1.12.0, etc)
   -e      Edition version (beta4, etc)
   -a      AMI ID (ami-123456, etc)
   -r      AMI source region (us-east-1)
   -c      Release Channel (beta, nightly, etc)
   -u      Cloud Release Channel (beta, nightly, etc)
   -l      AWS account list URL
   -p      Make AMI public (yes, no)
EOF
}

while getopts "hc:u:l:d:e:a:r:p:" OPTION
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
             EDITIONS_VERSION=$OPTARG
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
         u)
             CHANNEL_CLOUD=$OPTARG
             ;;
         l)
             AWS_ACCOUNT_LIST_URL=$OPTARG
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

if [[ -z $DOCKER_VERSION ]] || [[ -z $EDITIONS_VERSION ]] || [[ -z $AMI_ID ]] || [[ -z $AMI_SRC_REGION ]]
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
echo "EDITIONS_VERSION=$EDITIONS_VERSION"
echo "AMI_ID=$AMI_ID"
echo "AMI_SRC_REGION=$AMI_SRC_REGION"
echo "CHANNEL=$CHANNEL"
echo "CHANNEL_CLOUD=$CHANNEL_CLOUD"
echo "AWS_ACCOUNT_LIST_URL=$AWS_ACCOUNT_LIST_URL"
echo "MAKE_AMI_PUBLIC=$MAKE_AMI_PUBLIC"
echo "-------"
echo "== Prepare files =="

# prepare the files.
echo "== build docker image =="
# build the docker image
cd cfn_gen
docker build -t docker4x/cfngen -f Dockerfile .

echo "== run the docker image =="
echo "Using: EDITIONS_VERSION=$EDITIONS_VERSION \n AMI_ID=$AMI_ID \n AMI_SRC_REGION=$AMI_SRC_REGION \n CHANNEL=$CHANNEL"

# run the image
docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
-e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
-e DOCKER_VERSION=$DOCKER_VERSION \
-e EDITIONS_VERSION=$EDITIONS_VERSION \
-e AMI_ID=$AMI_ID \
-e AMI_SRC_REGION=$AMI_SRC_REGION \
-e CHANNEL="$CHANNEL" \
-e CHANNEL_CLOUD="$CHANNEL_CLOUD" \
-e AWS_ACCOUNT_LIST_URL="$AWS_ACCOUNT_LIST_URL" \
-e MAKE_AMI_PUBLIC="$MAKE_AMI_PUBLIC" \
-v `pwd`/cfn_gen/outputs:/home/docker/outputs \
docker4x/cfngen

# posthook

echo "== cleanup the docker image =="
# cleanup the image
docker rmi -f docker4x/cfngen

echo "== finished =="
echo "------"

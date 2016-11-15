#!/bin/bash

set -e
# run by calling ./run_release.sh docker_version edition_version ami_id ami_src_region
export PYTHONUNBUFFERED=1
# prehook

BUILD_NUMBER=${BUILD_NUMBER:-25}
DOCKER_VERSION=
EDITION_VERSION=
CHANNEL=

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
   -c      Release Channel (beta, nightly, etc)
EOF
}

while getopts "hc:d:e:" OPTION
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
         c)
             CHANNEL=$OPTARG
             ;;
         ?)
             usage
             exit
             ;;
     esac
done

if [[ -z $DOCKER_VERSION ]] || [[ -z $EDITION_VERSION ]]
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

AMI_LIST_URL="https://s3.amazonaws.com/docker-for-aws/data/ami/cs/$DOCKER_VERSION/ami_list.json"
AMI_LIST=$(curl -sf $AMI_LIST_URL)
export VERSION=aws-v$DOCKER_VERSION-$EDITION_VERSION-ddc

if [[ -z $AMI_LIST ]]
then
     echo "ERROR: There is no ami list available at $AMI_LIST_URL"
     echo "   This usually happens because an AMI hasn't been built for docker version: $DOCKER_VERSION yet."
     echo "   Run the `build-cs-ami.sh` and `release-cs-ami.sh` scripts to create and release the AMI"
     exit 1
fi

echo "------"
echo "Getting started with the release...."
echo "== Parameters =="
echo "BUILD_NUMBER=$BUILD_NUMBER"
echo "DOCKER_VERSION=$DOCKER_VERSION"
echo "EDITION_VERSION=$EDITION_VERSION"
echo "CHANNEL=$CHANNEL"
echo "VERSION=$VERSION"
echo "-------"
echo "== Prepare files =="

# prepare the files.
mkdir -p tmp
if [ -f tmp/docker_for_aws_ddc.template ]; then
    echo "Cleanup old template file."
    rm -f tmp/docker_for_aws_ddc.template
fi
echo "Copy over template file."
# need both since we use same image for both DDC and OSS and it looks for both
cp ../cloudformation/docker_for_aws.json tmp/docker_for_aws.template
cp ../cloudformation/docker_for_aws_ddc.json tmp/docker_for_aws_ddc.template


echo "=== Build editions docker images ==="
CURRPATH=`pwd`
# build binaries
# go to buoy dir
cd ../../tools/buoy/
./build_buoy.sh

cd ../metaserver/
./build.sh

echo "=== CURRPATH=$CURRPATH ==="
# back to release dir
cd $CURRPATH
# up to dockerfiles dir
cd ../dockerfiles/

# build images

./build_and_push_all.sh

echo "=== build controller ==="
cd files/elb-controller/container
DOCKER_TAG=$VERSION DOCKER_PUSH=true DOCKER_TAG_LATEST=false make -k container
cd $CURRPATH

# back to release dir
cd $CURRPATH

echo "== build docker image =="
# build the docker image
docker build -t docker4x/release-$BUILD_NUMBER -f Dockerfile.aws_release .

echo "== run the docker image =="
# run the image
docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
-e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
-e DOCKER_VERSION=$DOCKER_VERSION \
-e EDITION_VERSION=$EDITION_VERSION \
-e CHANNEL="$CHANNEL" \
docker4x/release-$BUILD_NUMBER /home/docker/ddc_release.sh

# posthook

echo "== cleanup the docker image =="
# cleanup the image
docker rmi -f docker4x/release-$BUILD_NUMBER

echo "== finished =="
echo "------"

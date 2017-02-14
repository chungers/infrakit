#!/bin/bash

set -e
# This script is expected to be running on an AWS instance with a correct IAM profile to build AMI's
# also it needs AWS creds, and a docker account that is logged in that has push permissions on docker4x domain.
# run by calling ./full_aws_release.sh -d docker_version -e edition_version -b docker_binary_url -c beta -l account_list_url
export PYTHONUNBUFFERED=1
# prehook

DOCKER_BIN_URL=
DOCKER_VERSION=
EDITION_VERSION=
CHANNEL=
MOBY_BRANCH="master"
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
   -d      Docker version (1.12.0, etc)
   -e      Edition version (beta4, etc)
   -b      Docker Bin URL (location where tar.gz file can be downloaded)
   -c      Release Channel (beta, nightly, etc)
   -l      AWS account list URL
   -p      Make AMI public (yes, no)
   -m      Moby Branch (master, 1.13.x, etc)
EOF
}

while getopts "hc:l:d:e:b:p:m:" OPTION
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
         b)
             DOCKER_BIN_URL=$OPTARG
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
         m)
             MOBY_BRANCH=$OPTARG
             ;;
         ?)
             usage
             exit
             ;;
     esac
done

if [[ -z $DOCKER_VERSION ]] || [[ -z $EDITION_VERSION ]] || [[ -z $DOCKER_BIN_URL ]]
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

export EDITIONS_VERSION=$DOCKER_VERSION-$EDITION_VERSION
export TAG_KEY=$EDITIONS_VERSION
export AMI_SRC_REGION=$(curl --silent http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r .region)
export HUB_LOGIN_ID=$(docker info | grep Username)

echo "------"
echo "Getting started with the release...."
echo "== Parameters =="
echo "BUILD_NUMBER=$BUILD_NUMBER"
echo "DOCKER_VERSION=$DOCKER_VERSION"
echo "EDITION_VERSION=$EDITION_VERSION"
echo "EDITIONS_VERSION=$EDITIONS_VERSION"
echo "DOCKER_BIN_URL=$DOCKER_BIN_URL"
echo "AMI_SRC_REGION=$AMI_SRC_REGION"
echo "CHANNEL=$CHANNEL"
echo "MOBY_BRANCH=$MOBY_BRANCH"
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
echo " /moby  <-- github.com/docker/moby "
echo "Make sure you run this script from the editions/aws/release directory."

MOBY_DIR=../../../moby/alpine
BASE_DIR=`pwd`

if [ ! -d "$MOBY_DIR" ]; then
    echo "$MOBY_DIR doesn't exist"
else
    echo "$MOBY_DIR Looks good!"
fi
echo "-------"
echo "== Build AMI =="
cd $MOBY_DIR
git checkout $MOBY_BRANCH
git pull
git clean -f -d

make ami-clean-mount || true
make clean || true
make ami DOCKER_BIN_URL=$DOCKER_BIN_URL
make ami-clean-mount
AMI_ID=$(cat ./cloud/aws/ami_id.out)
echo "AMI_ID=$AMI_ID"

# move out of the way, so it doesn't cause problems later.
mv -f ./cloud/aws/ami_id.out ./cloud/aws/ami_id.out.done

echo "== Build Docker images =="

cd $BASE_DIR

export VERSION=aws-v$DOCKER_VERSION-$EDITION_VERSION
echo "Version=$VERSION"

echo "= Build Buoy ="
cd ../../tools/buoy
./build_buoy.sh

echo "= Build Metaserver ="
cd ../metaserver
./build.sh

cd $BASE_DIR
cd ../dockerfiles

echo "= Build aws images ="
# build images

./build_and_push_all.sh

echo "= Build lb-container ="

cd files/elb-controller/container
DOCKER_TAG=$VERSION DOCKER_PUSH=true DOCKER_TAG_LATEST=false make -k container

echo "== Build Docker images =="
cd $BASE_DIR

# run release, this will create CFN templates and push them to s3, push AMI to different regions and share with list of approved accounts.
./new_run_release.sh -d $DOCKER_VERSION -e $EDITION_VERSION -a $AMI_ID -r $AMI_SRC_REGION -c $CHANNEL -l $DOCKER_AWS_ACCOUNT_URL -u cloud-$CHANNEL -p $MAKE_AMI_PUBLIC

echo "===== Done ====="

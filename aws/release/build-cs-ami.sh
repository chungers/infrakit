#!/bin/bash
set -e
# first parameter is url for docker binaries
# second parameter is if you want to make this ami the latest one.
# latest ones are used in nightly builds.
# example CS binary URL:
# https://packages.docker.com/1.12/builds/linux/amd64/docker-1.12.2-cs2.tgz
echo "================================================"
echo ""
echo ""
echo "Start CS AMI build"
echo $PATH
PATH=$PATH:/usr/local/bin
echo $PATH
export PYTHONUNBUFFERED=1

BIN_URL=$1
echo "BIN_URL is $BIN_URL"

if [ -n "$2" ]
then
 UPDATE_LATEST="True"
 echo "UPDATE_LATEST is True"
else
 UPDATE_LATEST="False"
fi


cd /tmp
curl -fsSL -o docker.tgz $BIN_URL
if [ ! -f docker.tgz ]; then
    echo "ERROR: No Docker tar file found on $BIN_URL check URL and try again."
    exit 1
fi
tar xzf docker.tgz
cd  docker
chmod 755 docker
export DOCKER_VERSION=$(./docker version -f '{{.Client.Version}}')
cd ..
rm -rf docker docker.tgz
echo "Docker version = $DOCKER_VERSION"

export DOCKER_EXPERIMENTAL=0
export TAG_KEY=$DOCKER_VERSION
export DOCKER_BIN_URL=$BIN_URL
MOBY_ROOT="/home/ubuntu/code/moby-master/alpine"

cd $MOBY_ROOT
git pull

# clean mount before we start to make sure.
make ami-clean-mount
make clean
make ami DOCKER_BIN_URL=$DOCKER_BIN_URL
make ami-clean-mount
# clean up after ourselves

echo "Finished AMI build, lets move onto next part."

# get ami-id
# look for anyfiles that look like ami_*.out those are ami's ready to be processed.
for f in ${MOBY_ROOT}/cloud/aws/ami_*.out; do

    ## Check if the glob gets expanded to existing files.
    ## If not, f here will be exactly the pattern above
    ## and the exists test will evaluate to false.
    if [ -e "$f" ]; then
        AMI_ID=$(cat $f)
        mv $f $f.done
    fi

    ## We can only process one at a time.
    break
done

if [[ -z $AMI_ID ]]
then
    echo "There is no AMI_ID, nothing to do, so stopping."
    # there are no AMI's ready skip.
     exit 1
fi
echo $AMI_ID

AMI_SRC_REGION=us-east-1
DOCKER_AWS_ACCOUNT_URL=https://s3.amazonaws.com/docker-for-aws/data/docker_accounts.txt

export AWS_ACCESS_KEY_ID=$(cat ~/.aws/credentials | grep aws_access_key_id | cut -f2 -d= | sed -e 's/^[ \t]*//')
export AWS_SECRET_ACCESS_KEY=$(cat ~/.aws/credentials | grep aws_secret_access_key | cut -f2 -d= | sed -e 's/^[ \t]*//')

cd /home/ubuntu/code/editions/aws/release

SH_IMAGE_NAME="docker4x/build-cs-ami"

echo "== build docker image =="
# build the docker image
docker build -t $SH_IMAGE_NAME -f Dockerfile.aws_release .

echo "== run the docker image =="

echo "== build AMI =="
# run the image
docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
-e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
-e DOCKER_VERSION=$DOCKER_VERSION \
-e AMI_ID=$AMI_ID \
-e AMI_SRC_REGION=$AMI_SRC_REGION \
-e UPDATE_LATEST="$UPDATE_LATEST" \
$SH_IMAGE_NAME /home/docker/build-cs-ami.sh

echo "== release AMI =="
# run the image
docker run -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
-e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
-e DOCKER_VERSION=$DOCKER_VERSION \
-e AMI_ID=$AMI_ID \
-e AMI_SRC_REGION=$AMI_SRC_REGION \
-e AWS_ACCOUNT_LIST_URL="$DOCKER_AWS_ACCOUNT_URL" \
$SH_IMAGE_NAME /home/docker/release-cs-ami.sh

echo "== cleanup the docker image =="
# cleanup the image
docker rmi -f $SH_IMAGE_NAME

echo "== finished =="
echo "------"

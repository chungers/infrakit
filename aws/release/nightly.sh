#!/bin/bash
set -e
echo "================================================"
echo ""
echo ""
echo "Start nightly build"
echo $PATH
PATH=$PATH:/usr/local/bin
echo $PATH
DAY=$(date +"%m_%d_%Y")
HASH=$(curl -f -s https://master.dockerproject.org/commit)
if [[ -z $HASH ]]
then
   echo "No valid HASH $HASH"
   HASH="NOHASHAVAILABLE"
fi
export AWS_ACCESS_KEY_ID=$(cat ~/.aws/credentials | grep aws_access_key_id | cut -f2 -d= | sed -e 's/^[ \t]*//')
export AWS_SECRET_ACCESS_KEY=$(cat ~/.aws/credentials | grep aws_secret_access_key | cut -f2 -d= | sed -e 's/^[ \t]*//')
export PYTHONUNBUFFERED=1
export CHANNEL="nightly"
BUILD_HOME="/home/ubuntu"
AMI_OUT_DIR="$BUILD_HOME/out"
AMI_OUT_FILE="ami_id_${DAY}.out"
export TAG_KEY="aws-nightly-${DAY}-${HASH}"

MASTER_DOCKER_VERSION=$(curl -s https://master.dockerproject.org/version)

export DOCKER_BIN_URL="https://master.dockerproject.org/linux/amd64/docker-$MASTER_DOCKER_VERSION.tgz"

cd $BUILD_HOME/code/moby/alpine
git checkout master
git pull
git clean -f -d

make clean
make ami-clean-mount
make ami DOCKER_BIN_URL=$DOCKER_BIN_URL
make ami-clean-mount
mv -f ./cloud/aws/ami_id.out $AMI_OUT_DIR/$AMI_OUT_FILE

echo "Finished AMI build, lets move onto next part."

# get ami-id
# look for anyfiles that look like ami_*.out those are ami's ready to be processed.
for f in $AMI_OUT_DIR/ami_*.out; do

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

AMI_SOURCE_REGION=us-east-1
DOCKER_AWS_ACCOUNT_URL=https://s3.amazonaws.com/docker-for-aws/data/docker_accounts.txt

# git update
cd $BUILD_HOME/code/editions/
git pull

# get docker version
export DOCKER_VERSION=$MASTER_DOCKER_VERSION
export EDITION_VERSION=nightly_$DAY
export VERSION=aws-v$DOCKER_VERSION-$EDITION_VERSION

# build binaries
cd $BUILD_HOME/code/editions/tools/buoy/
./build_buoy.sh

cd $BUILD_HOME/code/editions/tools/metaserver/
./build.sh

cd $BUILD_HOME/code/editions/aws/dockerfiles/

# build images

./build_and_push_all.sh

cd files/elb-controller/container
DOCKER_TAG=$VERSION DOCKER_PUSH=true DOCKER_TAG_LATEST=false make -k container
cd $CURRPATH

cd $BUILD_HOME/code/editions/aws/release

# run release
./run_release.sh -d $DOCKER_VERSION -e $EDITION_VERSION -a $AMI_ID -r $AMI_SOURCE_REGION -c nightly -l $DOCKER_AWS_ACCOUNT_URL -u cloud-nightly -p no

# run cleanup, remove things that are more than X days old.
python cleanup.py

# run s3_cleanup, remove buckets left over from DDC testing.
python s3_cleanup.py

# run tests
python test_cfn.py -c https://docker-for-aws.s3.amazonaws.com/aws/nightly/latest.json -f results -t oss
python test_cfn.py -c https://docker-for-aws.s3.amazonaws.com/aws/cloud-nightly/latest.json -f cloud_results -t cloud

# Rebuild the nightly index page.
python build_index.py

# notify results
python notify.py

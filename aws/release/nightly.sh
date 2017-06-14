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

AMI_BUCKET="docker-ci-editions"
AMI_URL="ami/ami_id.out"

# git update
cd $BUILD_HOME/code/editions/
git pull

# get docker version
EDITIONS_META=$(aws s3api --no-sign-request get-object --bucket $AMI_BUCKET --key ${AMI_URL} docker.out | jq -r '.Metadata')
export EDITIONS_VERSION=$(echo $EDITIONS_META | jq -r '.editions_version')
export DOCKER_VERSION=$(echo $EDITIONS_META | jq -r '.docker_version')
export MOBY_COMMIT=$(echo $EDITIONS_META | jq -r '.moby_commit')
export AWS_TARGET_PATH="dist/aws/$CHANNEL/$AWS_TAG_VERSION"

export RELEASE=true

AMI_SOURCE_REGION=us-west-2

mkdir -p $AMI_OUT_DIR
# Download the ami_id.out
aws s3 --no-sign-request cp s3://${AMI_BUCKET}/${AMI_URL} $AMI_OUT_DIR/ami_id.out
echo "AMI build captured, lets move onto next part."

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
echo "AMI: $AMI_ID is available in $AMI_SRC_REGION"


DOCKER_AWS_ACCOUNT_URL=https://s3.amazonaws.com/docker-for-aws/data/docker_accounts.txt

cd $BUILD_HOME/code/editions/aws/release

# run release
./run_release.sh -d $DOCKER_VERSION -e $EDITIONS_VERSION -a $AMI_ID -r $AMI_SOURCE_REGION -c nightly -l $DOCKER_AWS_ACCOUNT_URL -u cloud-nightly -p no

# run cleanup, remove things that are more than X days old.
python cleanup.py

# run s3_cleanup, remove buckets left over from DDC testing.
python ../common/s3_cleanup.py

# sleep to help with API throttle limits
sleep 60

# run tests
python ../test/cfn.py -c https://docker-for-aws.s3.amazonaws.com/aws/nightly/latest.json -f results -t oss

# sleep to help with API throttle limits
sleep 60
python ../test/cfn.py -c https://docker-for-aws.s3.amazonaws.com/aws/cloud-nightly/latest.json -f cloud_results -t cloud

# Rebuild the nightly index page.
python build_index.py

# notify results
python notify.py

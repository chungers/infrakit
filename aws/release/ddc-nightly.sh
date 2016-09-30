#!/bin/bash
set -e
echo "================================================"
echo ""
echo ""
echo "Start nightly for ddc"
echo $PATH
PATH=$PATH:/usr/local/bin
echo $PATH

export PYTHONUNBUFFERED=1

NOW=$(date +"%m_%d_%Y")
export EDITION_VERSION=nightly_$NOW
export DOCKER_VERSION=$(curl -fs https://s3.amazonaws.com/docker-for-aws/data/ami/cs/latest.txt)
export VERSION=aws-v$DOCKER_VERSION-$EDITION_VERSION-ddc

export AWS_ACCESS_KEY_ID=$(cat ~/.aws/credentials | grep aws_access_key_id | cut -f2 -d= | sed -e 's/^[ \t]*//')
export AWS_SECRET_ACCESS_KEY=$(cat ~/.aws/credentials | grep aws_secret_access_key | cut -f2 -d= | sed -e 's/^[ \t]*//')

# git update
cd /home/ubuntu/code/editions/
git pull

cd /home/ubuntu/code/editions/aws/release

# run release
./run_ddc_release.sh -d $DOCKER_VERSION -e $EDITION_VERSION -c ddc-nightly

# # run tests
python test_cfn.py -c https://docker-for-aws.s3.amazonaws.com/aws/ddc-nightly/latest.json -f ddc_results -t ddc

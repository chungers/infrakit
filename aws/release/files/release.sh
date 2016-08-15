#!/bin/bash

# requires the following ENV Variables to be set.
# AWS_ACCESS_KEY_ID
# AWS_SECRET_ACCESS_KEY
# DOCKER_VERSION
# EDITION_VERSION
# AMI_ID
# AMI_SRC_REGION

# don't buffer the output
export PYTHONUNBUFFERED=1

python /home/docker/aws_release.py --docker_version="$DOCKER_VERSION" \
    --edition_version="$EDITION_VERSION" --ami_id="$AMI_ID" \
    --ami_src_region="$AMI_SRC_REGION" --channel="$CHANNEL" --account_list_url="$AWS_ACCOUNT_LIST_URL"

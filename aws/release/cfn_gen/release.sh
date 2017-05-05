#!/bin/bash

# requires the following ENV Variables to be set.
# AWS_ACCESS_KEY_ID
# AWS_SECRET_ACCESS_KEY
# DOCKER_VERSION
# EDITION_TAG
# AMI_ID
# AMI_SRC_REGION
# MAKE_AMI_PUBLIC

# don't buffer the output
export PYTHONUNBUFFERED=1



python /home/docker/release.py --docker_version="$DOCKER_VERSION" \
    --editions_version="$EDITIONS_VERSION" --ami_id="$AMI_ID" \
    --ami_src_region="$AMI_SRC_REGION" --channel="$CHANNEL" \
    --channel_cloud="$CHANNEL_CLOUD" --account_list_url="$AWS_ACCOUNT_LIST_URL" \
    --public="$MAKE_AMI_PUBLIC" $FLAGS

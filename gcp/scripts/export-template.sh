#!/bin/bash

set -ex

export CLOUDSDK_CORE_PROJECT='docker-for-gcp'

BUCKET="gs://docker-for-gcp-templates"

gsutil ls -b ${BUCKET} || gsutil mb ${BUCKET}
gsutil -m rsync -r ../templates/ ${BUCKET}
gsutil cp get.sh ${BUCKET}
gsutil -m acl -r set public-read ${BUCKET}
gsutil -m setmeta -h "Cache-Control:private, max-age=0, no-transform" -r ${BUCKET}

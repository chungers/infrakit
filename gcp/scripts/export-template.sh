#!/bin/bash

set -ex

export CLOUDSDK_CORE_PROJECT='code-story-blog'
export CLOUDSDK_COMPUTE_ZONE='europe-west1-d'

BUCKET="gs://docker-template"

gsutil ls -b ${BUCKET} || gsutil mb ${BUCKET}
gsutil -m rsync -r ../configuration/ ${BUCKET}
gsutil cp get.sh ${BUCKET}
gsutil -m acl -r set public-read ${BUCKET}
gsutil -m setmeta -h "Cache-Control:private, max-age=0, no-transform" -r ${BUCKET}

#!/bin/bash

set -ex

export CLOUDSDK_COMPUTE_ZONE='europe-west1-d'
export CLOUDSDK_CORE_PROJECT='code-story-blog'
VM='exportdisk'

echo "Cleaning up"
gcloud compute instances describe ${VM} >/dev/null 2>&1 && (yes | gcloud compute instances delete ${VM} --delete-disks=all)

echo "Create a machine to do the export"
gcloud compute disks create image-disk --image-project=${CLOUDSDK_CORE_PROJECT} --image docker
gcloud compute instances create ${VM} \
	--machine-type "n1-standard-4" \
	--scopes storage-rw \
  --image "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-8-jessie-v20161027" \
  --boot-disk-size "500GB" \
  --boot-disk-device-name "${VM}" \
  --boot-disk-type "pd-ssd"
gcloud compute instances attach-disk ${VM} --disk="image-disk" --device-name="image-disk" --mode=rw

echo "Export the image to a file and upload it to Google Cloud Storage"
gcloud compute ssh ${VM} -- -o ConnectionAttempts=30 -o ConnectTimeout=10 'bash -s' << EOF
sudo dd if=/dev/disk/by-id/google-image-disk of=/tmp/disk.raw bs=4M conv=sparse
cd /tmp
sudo tar czSf docker.image.tar.gz disk.raw
gsutil mb gs://docker-image
gsutil cp docker.image.tar.gz gs://docker-image
gsutil acl set public-read gs://docker-image/docker.image.tar.gz
EOF

echo "Cleaning up"
yes | gcloud compute instances delete ${VM} --delete-disks=all

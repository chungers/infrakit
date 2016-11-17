#!/bin/bash

set -ex

export CLOUDSDK_COMPUTE_ZONE='europe-west1-d'
export CLOUDSDK_CORE_PROJECT='code-story-blog'

VM='exportdisk'

echo "Clean up"
gcloud compute instances describe ${VM} >/dev/null 2>&1 && gcloud -q compute instances delete ${VM} --delete-disks=all

echo "Create a short-lived instance to do the export"
gcloud compute disks create image-disk --image docker-image
gcloud compute instances create ${VM} \
	--machine-type "n1-highcpu-4" \
	--scopes storage-rw \
	--image-family=debian-8 \
	--image-project=debian-cloud \
	--boot-disk-size "500GB" \
  --boot-disk-type "pd-ssd"
gcloud compute instances attach-disk ${VM} --disk="image-disk" --device-name="image-disk"

echo "Export the image to a file and upload it to Google Cloud Storage"
gcloud compute ssh ${VM} -- -o ConnectionAttempts=30 -o ConnectTimeout=10 'bash -s' << EOF
set -x
sudo mkdir /mnt/image-disk
sudo mount /dev/disk/by-id/google-image-disk-part1 /mnt/image-disk
sudo rm -Rf /mnt/image-disk/home/*
echo '{"log-driver":"gcplogs"}' | sudo tee /mnt/image-disk/etc/docker/daemon.json
sudo umount /mnt/image-disk/
sudo dd if=/dev/disk/by-id/google-image-disk of=/tmp/disk.raw bs=4M conv=sparse
cd /tmp
sudo apt-get install -y pigz
sudo tar --use-compress-program=pigz -cSf docker.image.tar.gz disk.raw
gsutil mb gs://docker-image
gsutil -h "Cache-Control:private, max-age=0, no-transform" cp -a public-read docker.image.tar.gz gs://docker-image
EOF

echo "Cleaning up"
gcloud -q compute instances delete ${VM} --delete-disks=all

#!/bin/bash

set -ex

function metadata {
  curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/${1}
}

# Regenerates SSH host keys when the VM is restarted with a new IP address.
# Booting a VM from an image with a known SSH key allows a number of attacks.
# This function will regenerating the host key whenever the IP address
# changes. This applies the first time the instance is booted, and each time
# the disk is used to boot a new instance.
# See https://github.com/GoogleCloudPlatform/compute-image-packages/blob/master/google_compute_engine/instance_setup/instance_setup.py#L133

# Read the instance id from the metadata server
INSTANCE_ID=$(metadata 'instance/id')

# Read the previous instance id saved on disk
if [ -f /etc/instance_id ]; then
  PREVIOUS_INSTANCE_ID=$(cat /etc/instance_id)
fi

# Regenerate the ssh keys if it doesn't match
if [ "${INSTANCE_ID}" != "${PREVIOUS_INSTANCE_ID}" ]; then
  ssh-keygen -A
  echo ${INSTANCE_ID} > /etc/instance_id
fi

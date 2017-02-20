#!/bin/sh

export CLOUDSDK_CORE_PROJECT={{ PROJECT }}
export CLOUDSDK_COMPUTE_ZONE={{ ZONE }}

curl -sfH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/attributes/manager-script > /start.sh

echo Wait for the base disk image
while :; do gcloud compute images describe {{ STACK }}-disk-image-{{ VERSION }} && break || sleep 1; done

echo Wait for the network
while :; do gcloud compute networks describe {{ STACK }}-network && break || sleep 1; done

gcloud compute disks describe {{ STACK }}-manager-1
if [ $? -eq 0 ]; then
  gcloud compute instances create {{ STACK }}-manager-1 \
    --machine-type {{ properties['managerMachineType'] }} \
    --network projects/{{ PROJECT }}/global/networks/{{ STACK }}-network \
    --tags swarm,swarm-manager \
    --scopes https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring.write,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/trace.append,https://www.googleapis.com/auth/compute,https://www.googleapis.com/auth/cloudruntimeconfig \
    --metadata infrakit--group={{ STACK }}-managers,infrakit--config_sha=bootstrap \
    --metadata-from-file startup-script=/start.sh \
    --disk=boot=yes,device-name={{ STACK }}-manager-1,name={{ STACK }}-manager-1
else
  gcloud compute instances create {{ STACK }}-manager-1 \
    --machine-type {{ properties['managerMachineType'] }} \
    --network projects/{{ PROJECT }}/global/networks/{{ STACK }}-network \
    --image projects/{{ PROJECT }}/global/images/{{ STACK }}-disk-image-{{ VERSION }} \
    --tags swarm,swarm-manager \
    --scopes https://www.googleapis.com/auth/devstorage.read_only,https://www.googleapis.com/auth/logging.write,https://www.googleapis.com/auth/monitoring.write,https://www.googleapis.com/auth/service.management.readonly,https://www.googleapis.com/auth/servicecontrol,https://www.googleapis.com/auth/trace.append,https://www.googleapis.com/auth/compute,https://www.googleapis.com/auth/cloudruntimeconfig \
    --metadata infrakit--group={{ STACK }}-managers,infrakit--config_sha=bootstrap \
    --metadata-from-file startup-script=/start.sh \
    --boot-disk-size {{ properties['managerDiskSize'] }}GB \
    --boot-disk-type {{ properties['managerDiskType'] }} \
    --boot-disk-device-name {{ STACK }}-manager-1 \
    --no-boot-disk-auto-delete
fi

echo Wait for the target pool
while :; do gcloud compute target-pools describe --region={{REGION}} {{ STACK }}-lb-pool && break || sleep 1; done

gcloud compute target-pools add-instances \
  {{ STACK }}-lb-pool \
  --instances={{ STACK }}-manager-1 \
  --instances-zone {{ ZONE }}

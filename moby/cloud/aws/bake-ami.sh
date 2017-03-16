#!/bin/sh

# Script to automate creation and snapshotting of a Moby AMI.  Currently, it's
# intended to be invoked from an instance running in the same region as the
# target AMI will be in, since it directly mounts the created EBS volume as a
# device on this running instance.

set -e

PROVIDER="aws"

. "./build-common.sh"
. "./common.sh"

export AWS_DEFAULT_REGION=$(current_instance_region)

# TODO(nathanleclaire): This device could be calculated dynamically to avoid conflicts.
EBS_DEVICE=/dev/xvdb
CHANNEL=${CHANNEL:-editions}
DAY=$(date +"%m_%d_%Y")

bake_image()
{
	# Create a new EBS volume.  We will format this volume to boot into Moby
	# initrd via syslinux in MBR.  That formatted drive can then be snapshotted
	# and turned into an AMI.
	VOLUME_ID=$(aws ec2 create-volume \
		--size 1 \
		--availability-zone $(current_instance_az) | jq -r .VolumeId)

	tag ${VOLUME_ID} ${DAY} ${CHANNEL}

	aws ec2 wait volume-available --volume-ids ${VOLUME_ID}

	arrowecho "Attaching volume"
	aws ec2 attach-volume \
		--volume-id ${VOLUME_ID} \
		--device ${EBS_DEVICE} \
		--instance-id $(current_instance_id) >/dev/null

	aws ec2 wait volume-in-use --volume-ids ${VOLUME_ID}

	format_on_device "${EBS_DEVICE}"
	configure_syslinux_on_device_partition "${EBS_DEVICE}" "${EBS_DEVICE}1"

	arrowecho "Taking snapshot!"

	# Take a snapshot of the volume we wrote to.
	SNAPSHOT_ID=$(aws ec2 create-snapshot \
		--volume-id ${VOLUME_ID} \
		--description "Snapshot of Moby device for AMI baking" | jq -r .SnapshotId)

	tag ${SNAPSHOT_ID} ${DAY} ${CHANNEL}

	arrowecho "Waiting for snapshot completion"

	aws ec2 wait snapshot-completed --snapshot-ids ${SNAPSHOT_ID}

	# Convert that snapshot into an AMI as the root device.
	IMAGE_ID=$(aws ec2 register-image \
		--name "${IMAGE_NAME}" \
		--description "${IMAGE_DESCRIPTION}" \
		--architecture x86_64 \
		--root-device-name "${EBS_DEVICE}" \
		--virtualization-type "hvm" \
		--block-device-mappings "[
			{
				\"DeviceName\": \"${EBS_DEVICE}\",
				\"Ebs\": {
					\"SnapshotId\": \"${SNAPSHOT_ID}\"
				}
			}
		]" | jq -r .ImageId)

	tag ${IMAGE_ID} ${DAY} ${CHANNEL}

	# Boom, now you (should) have a Moby AMI.
	arrowecho "Created AMI: ${IMAGE_ID}"

	if [ $LOAD_IMAGES == 'true' ]; then
		echo "${IMAGE_ID}" >"${MOBY_OUTPUT}/ami_id_ee.out"
	else
		echo "${IMAGE_ID}" >"${MOBY_OUTPUT}/ami_id.out"
	fi
}

clean_volume_mount()
{
	VOLUME_ID=$(aws ec2 describe-volumes --filters "Name=tag-key,Values=editions_version" "Name=tag-value,Values=$1" | jq -r .Volumes[0].VolumeId)
	if [ ${VOLUME_ID} = "null" ]
	then
		arrowecho "No volume found, skipping"
	else
		arrowecho "Detaching volume"
		aws ec2 detach-volume --volume-id ${VOLUME_ID} >/dev/null || errecho "WARN: Error detaching volume!"
		aws ec2 wait volume-available --volume-ids ${VOLUME_ID}
		arrowecho "Deleting volume"
		aws ec2 delete-volume --volume-id ${VOLUME_ID} >/dev/null
	fi
}

clean_tagged_resources()
{
	clean_volume_mount $1

	IMAGE_ID=$(aws ec2 describe-images --filters "Name=tag-key,Values=editions_version" "Name=tag-value,Values=$1" | jq -r .Images[0].ImageId)
	if [ ${IMAGE_ID} = "null" ]
	then
		arrowecho "No image found, skipping"
	else
		arrowecho "Deregistering previously baked AMI"

		# Sometimes describe-images does not return null even if the found
		# image cannot be deregistered
		#
		# TODO(nathanleclaire): More elegant solution?
		aws ec2 deregister-image --image-id ${IMAGE_ID} >/dev/null || errecho "WARN: Issue deregistering previously tagged image!"
	fi

	SNAPSHOT_ID=$(aws ec2 describe-snapshots --filters "Name=tag-key,Values=editions_version" "Name=tag-value,Values=$1" | jq -r .Snapshots[0].SnapshotId)
	if [ ${SNAPSHOT_ID} = "null" ]
	then
		arrowecho "No snapshot found, skipping"
	else
		arrowecho "Deleting volume snapshot"
		aws ec2 delete-snapshot --snapshot-id ${SNAPSHOT_ID}
	fi
}

if [ -z "${AWS_ACCESS_KEY_ID}" ] || [ -z "${AWS_SECRET_ACCESS_KEY}" ]
then
	errecho "Must set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY to authenticate with AWS."
	exit 1
fi

case "$1" in
	bake)
		bake_image
		;;
	clean)
		arrowecho "Cleaning resources from previous build tag (${TAG_KEY_PREV}) if applicable..."
		clean_tagged_resources "${TAG_KEY_PREV}"
		arrowecho "Cleaning resources from current build tag (${TAG_KEY}) if applicable..."
		clean_tagged_resources "${TAG_KEY}"
		;;
	clean-mount)
		clean_volume_mount "${TAG_KEY}"
		;;
	*)
		errecho "Command $1 not found.  Usage: ./bake-ami.sh [bake|clean|clean-mount]"
esac

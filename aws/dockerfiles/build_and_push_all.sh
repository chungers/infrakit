#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
VERSION="${VERSION:-latest}"

echo "+ Creating dist folder: $AWS_TARGET_PATH"
mkdir -p $ROOTDIR/$AWS_TARGET_PATH

for IMAGE in shell init guide ddc-init cloud meta
do
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-aws:${VERSION}"
	docker build --pull -t "${FINAL_IMAGE}" -f "Dockerfile.${IMAGE}" .
	docker save "${FINAL_IMAGE}" > "${ROOTDIR}/${AWS_TARGET_PATH}/${IMAGE}-aws.tar"
	if [ $RELEASE -eq 1 ]; then
		docker push "${FINAL_IMAGE}"
	fi
done

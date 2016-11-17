#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE="${NAMESPACE:-docker4x}"
VERSION="${VERSION:-latest}"

for IMAGE in shell init guide ddc-init cloud meta
do
	FINAL_IMAGE="${NAMESPACE}/${IMAGE}-aws:${VERSION}"
	docker build --pull -t "${FINAL_IMAGE}" -f "Dockerfile.${IMAGE}" .
	docker push "${FINAL_IMAGE}"
done

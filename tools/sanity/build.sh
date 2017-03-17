#!/bin/bash

set -e

NAMESPACE="${NAMESPACE:-docker4x}"
VERSION="${VERSION:-latest}"
FINAL_IMAGE="${NAMESPACE}/sanity:${VERSION}"

docker build --pull -t "${FINAL_IMAGE}" -f "Dockerfile" .
if [ ${DOCKER_PUSH} -eq 1 ]; then
	docker push "${FINAL_IMAGE}"
fi

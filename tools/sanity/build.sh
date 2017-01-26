#!/bin/bash

set -e

NAMESPACE="${NAMESPACE:-docker4x}"
VERSION="${VERSION:-latest}"
FINAL_IMAGE="${NAMESPACE}/sanity:${VERSION}"

docker build --pull -t "${FINAL_IMAGE}" -f "Dockerfile" .
docker push "${FINAL_IMAGE}"

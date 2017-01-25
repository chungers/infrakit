#!/bin/bash

set -ex

INFRAKIT_IMAGE="infrakit/devbundle:master-1116"
INFRAKIT_GCP_IMAGE="infrakit/gcp:master-55"

mkdir -p bin

# Extract Infrakit binaries
docker rm -f infrakit-build | true
docker run --name=infrakit-build ${INFRAKIT_IMAGE} true

BINARIES="infrakit-flavor-combo infrakit-flavor-swarm infrakit-flavor-vanilla infrakit-group-default infrakit-manager infrakit"
for BINARY in $BINARIES; do
  docker cp infrakit-build:/usr/local/bin/${BINARY} bin
done

# Extract Infrakit.gcp binaries
docker rm -f infrakit-build | true
docker run --name=infrakit-build ${INFRAKIT_GCP_IMAGE} true

BINARIES="infrakit-instance-gcp"
for BINARY in $BINARIES; do
  docker cp infrakit-build:/usr/local/bin/${BINARY} bin
done

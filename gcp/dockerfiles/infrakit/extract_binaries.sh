#!/bin/bash

set -ex

INFRAKIT_IMAGE="infrakit/devbundle:master-1300"
INFRAKIT_GCP_IMAGE="infrakit/gcp:master-90"

mkdir -p bin

# Extract Infrakit binaries
docker container rm -f infrakit-build | true
docker container run --name=infrakit-build ${INFRAKIT_IMAGE} true

BINARIES="infrakit-flavor-swarm infrakit-flavor-vanilla infrakit-flavor-combo infrakit-group-default infrakit-manager infrakit"
for BINARY in $BINARIES; do
  docker container cp infrakit-build:/usr/local/bin/${BINARY} bin
done

# Extract Infrakit.gcp binaries
docker container rm -f infrakit-build | true
docker container run --name=infrakit-build ${INFRAKIT_GCP_IMAGE} true

BINARIES="infrakit-instance-gcp"
for BINARY in $BINARIES; do
  docker container cp infrakit-build:/usr/local/bin/${BINARY} bin
done

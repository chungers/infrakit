#!/bin/bash
set -e

# if there is an ENV with this name, use it, if not, default to these values.
NAMESPACE=${NAMESPACE:-docker4x}
VERSION=${VERSION:-azure-v1.12.1-rc1-beta5}

docker build -t $NAMESPACE/init-azure:$VERSION -f Dockerfile.init .
docker push $NAMESPACE/init-azure:$VERSION

docker build -t $NAMESPACE/guide-azure:$VERSION -f Dockerfile.guide .
docker push $NAMESPACE/guide-azure:$VERSION

# Helper container that creates the Active Directory Service Principal for API access.
docker build -t $NAMESPACE/create-sp-azure:$VERSION -t $NAMESPACE/create-sp-azure:latest -f Dockerfile.create-sp-azure .
docker push $NAMESPACE/create-sp-azure:$VERSION
docker push $NAMESPACE/create-sp-azure:latest

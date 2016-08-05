#!/bin/bash
set -e

NAMESPACE=docker4x
VERSION=azure-v1.12.0-beta4

docker build -t $NAMESPACE/init-azure:$VERSION -f Dockerfile.init .
docker push $NAMESPACE/init-azure:$VERSION

docker build -t $NAMESPACE/guide-azure:$VERSION -f Dockerfile.guide .
docker push $NAMESPACE/guide-azure:$VERSION

# Helper container that creates the Active Directory Service Principal for API access.
docker build -t $NAMESPACE/create-sp-azure:$VERSION -t $NAMESPACE/create-sp-azure:latest -f Dockerfile.create_sp .
docker push $NAMESPACE/create-sp-azure:$VERSION 
docker push $NAMESPACE/create-sp-azure:latest

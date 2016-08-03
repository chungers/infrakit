#!/bin/bash
set -e

NAMESPACE=docker4x
VERSION=aws-v1.12.0-beta4

docker build -t $NAMESPACE/shell-aws:$VERSION -f Dockerfile.shell .
docker push $NAMESPACE/shell-aws:$VERSION

docker build -t $NAMESPACE/init-aws:$VERSION -f Dockerfile.init .
docker push $NAMESPACE/init-aws:$VERSION

docker build -t $NAMESPACE/guide-aws:$VERSION -f Dockerfile.guide .
docker push $NAMESPACE/guide-aws:$VERSION

docker build -t $NAMESPACE/ddc-init-aws:$VERSION -f Dockerfile.ddc-init .
docker push $NAMESPACE/ddc-init-aws:$VERSION

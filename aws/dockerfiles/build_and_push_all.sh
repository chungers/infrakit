#!/bin/bash
NAMESPACE=docker4x
VERSION=aws-v1.12.0-rc3-beta2

docker build -t $NAMESPACE/shell-aws:$VERSION -f Dockerfile.shell .
docker push $NAMESPACE/shell-aws:$VERSION

docker build -t $NAMESPACE/init-aws:$VERSION -f Dockerfile.init .
docker push $NAMESPACE/init-aws:$VERSION

docker build -t $NAMESPACE/watchdog-aws:$VERSION -f Dockerfile.watchdog .
docker push $NAMESPACE/watchdog-aws:$VERSION

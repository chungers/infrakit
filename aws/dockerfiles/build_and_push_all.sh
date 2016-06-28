#!/bin/bash
VERSION=0.1.0

docker build -t docker4x/shell-aws:$VERSION -f Dockerfile.shell .
docker push docker4x/shell-aws:$VERSION

docker build -t docker4x/init-aws:$VERSION -f Dockerfile.init .
docker push docker4x/init-aws:$VERSION

docker build -t docker4x/watchdog-aws:$VERSION -f Dockerfile.watchdog .
docker push docker4x/watchdog-aws:$VERSION

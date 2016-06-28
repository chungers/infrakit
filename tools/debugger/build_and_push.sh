#!/bin/bash
VERSION=0.1
docker build -t docker4x/debugger:$VERSION -f Dockerfile.debugger.aws .
docker push docker4x/debugger:$VERSION

docker build -t docker4x/node-debug:$VERSION -f Dockerfile.node-debug.aws .
docker push docker4x/node-debug:$VERSION

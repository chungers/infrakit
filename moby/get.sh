#!/bin/sh

set -e

mkdir -p build/aws build/azure build/gcp src

docker rm moby || true
docker create --name moby mobylinux/media:${MOBY_IMG_COMMIT} true
docker cp moby:/initrd.img src/initrd.img
docker cp moby:/vmlinuz64 build/
docker rm moby

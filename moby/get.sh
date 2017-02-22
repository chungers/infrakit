#!/bin/sh

set -e 

rm -rf build
mkdir -p build

docker create --name moby mobylinux/media:aufs-$MOBY_IMG_COMMIT ls
docker cp moby:/initrd.img build/
docker cp moby:/vmlinuz64 build/
docker rm moby
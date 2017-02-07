#!/bin/bash

# Only used for testing, will be removed once testing is complete.

docker build -t docker4x/cfnbuild -f Dockerfile .

docker run -v `pwd`/outputs:/home/docker/outputs docker4x/cfnbuild

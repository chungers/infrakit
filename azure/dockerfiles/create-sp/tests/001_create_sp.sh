#!/bin/sh

set -ex

echo Check that create-sp binary is working properly
ENV=test /usr/bin/create_sp.sh 2>&1 | grep -q "usage: docker run -ti docker4x/create-sp-azure"

#!/bin/sh

set -ex

echo Check that metaserver binary is working properly
/usr/bin/metaserver --help 2>&1 | grep -q "Usage of /usr/bin/metaserver"

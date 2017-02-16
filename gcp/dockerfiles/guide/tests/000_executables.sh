#!/bin/sh

set -ex

echo Check that all executable files are there
test -x /usr/bin/buoy.sh
test -x /usr/bin/vacuum.sh
test -x /usr/bin/cleanup.sh
test -x /usr/bin/buoy

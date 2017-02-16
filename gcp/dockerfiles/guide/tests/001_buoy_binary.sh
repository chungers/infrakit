#!/bin/sh

set -ex

echo Check that buoy binary is working properly
/usr/bin/buoy --help 2>&1 | grep -q "Usage of /usr/bin/buoy"

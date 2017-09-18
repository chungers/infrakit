#!/bin/sh

set -ex

echo Check that all executable files are there
test -x /usr/bin/buoy
test -x /usr/bin/docker
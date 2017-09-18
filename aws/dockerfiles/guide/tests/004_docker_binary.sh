#!/bin/sh

set -ex

echo Check that docker binary is working properly
/usr/bin/docker --version 2>&1 | grep -q "Docker version"

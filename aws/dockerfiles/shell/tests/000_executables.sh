#!/bin/sh

set -ex

echo Check that all executable files are there
test -x /usr/bin/docker-diagnose
test -x /usr/bin/swarm-exec
test -x /usr/bin/docker
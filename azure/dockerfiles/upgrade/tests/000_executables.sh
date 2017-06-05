#!/bin/sh

set -ex

echo "Check that all executable files are there"
test -x /usr/bin/upgrade.sh 
test -x /opt/azupgrade.py

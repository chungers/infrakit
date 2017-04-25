#!/bin/sh

set -ex

echo Check that all executable files are there
test -x /entry.sh
test -x /usr/bin/azureleader.py 
test -x /usr/bin/sakey.py
test -x /usr/bin/aztags.py 
test -x /usr/bin/buoy
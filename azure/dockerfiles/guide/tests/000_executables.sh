#!/bin/sh

set -ex

echo Check that all executable files are there
test -x /entry.sh
test -x /usr/bin/azparameters.py 
test -x /usr/bin/vacuum.sh
test -x /usr/bin/buoy 
test -x /usr/bin/buoy.sh 
test -x /usr/bin/refresh.sh
test -x /usr/bin/azutils.py 
test -x /usr/bin/azupgrade.py
test -x /usr/bin/azupg_listener.py 
test -x /usr/bin/kick_azupg_listener.sh
test -x /usr/bin/azrejoin.py 
test -x /usr/bin/kick_azrejoin_listener.sh
test -x /usr/bin/aznodewatch.py 
test -x /usr/bin/kick_aznodewatch.sh
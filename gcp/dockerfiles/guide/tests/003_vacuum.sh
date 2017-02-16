#!/bin/sh

set -ex

echo Check that vacuum script can run
/usr/bin/vacuum.sh

echo Check that vacuum script can run
SLEEP=no RUN_VACUUM=yes /usr/bin/vacuum.sh

#!/bin/sh

set -ex

echo Check that supervisord binary is working properly
/usr/bin/supervisord --help 2>&1 | grep -q "Usage: /usr/bin/supervisord"
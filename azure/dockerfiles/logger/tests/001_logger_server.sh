#!/bin/sh

set -ex

echo Check that log server is working properly
/server.py 2>&1 | grep -q "KeyError: 'ACCOUNT_ID'"

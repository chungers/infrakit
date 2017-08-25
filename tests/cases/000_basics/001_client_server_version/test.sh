#!/bin/sh
# SUMMARY: Checks that the server and client Docker versions match
# LABELS:

set -e

. "${RT_PROJECT_ROOT}/_lib/lib.sh"

[ $(docker version --format '{{.Client.Version}}') = $(docker version --format '{{.Server.Version}}') ]

exit 0
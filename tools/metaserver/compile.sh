#!/bin/bash

set -e

cd /go/src/metaserver
go build
cp metaserver /go/bin

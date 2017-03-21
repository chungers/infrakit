#!/bin/bash

set -e

cd /go/src/metaserver
go vet
go build
cp metaserver /go/bin

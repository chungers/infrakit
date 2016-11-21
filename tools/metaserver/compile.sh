#!/bin/bash

set -e

cd /go/src/metaserver
trash
go vet
go build
cp metaserver /go/bin

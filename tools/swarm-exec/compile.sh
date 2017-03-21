#!/bin/bash

set -e

cd /go/src/swarm-exec
go vet
go build -tags netgo
cp swarm-exec /go/bin

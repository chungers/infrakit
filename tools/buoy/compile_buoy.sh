#!/bin/bash

set -e

cd /go/src/buoy
go build
cp buoy /go/bin

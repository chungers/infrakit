#!/bin/sh

set -ex

echo Check that buoy binary is working properly
/registration --help 2>&1 | grep -q "Docker Swarm Cluster Registration"
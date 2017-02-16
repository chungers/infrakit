#!/bin/sh

set -ex

echo Check that buoy script is working properly
CHANNEL=test NODE_TYPE=worker /usr/bin/buoy.sh node:join
CHANNEL=test NODE_TYPE=manager /usr/bin/buoy.sh node:join
CHANNEL=test NODE_TYPE=leader /usr/bin/buoy.sh identify
CHANNEL=test NODE_TYPE=leader /usr/bin/buoy.sh swarm:init

#!/bin/bash
/usr/bin/docker run -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker -v /var/lib/waagent/CustomData:/var/lib/waagent/CustomData -ti docker4x/upgrade-azure-core:$TAG

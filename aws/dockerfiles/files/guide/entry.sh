#!/bin/sh

# set system wide env variables, so they are available to ssh connections
/usr/bin/env > /etc/environment

# start cron
/usr/sbin/crond -f -L 8

#!/bin/sh

# set system wide env variables, so they are available to ssh connections
/usr/bin/env > /etc/environment

# start cron
/usr/sbin/crond -f -l 9 -L /var/log/cron.log

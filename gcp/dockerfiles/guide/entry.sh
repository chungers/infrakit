#!/bin/sh

echo "Initialize logging for guide daemons"
# setup symlink to output logs from relevant scripts to container logs
ln -s /proc/1/fd/1 /var/log/docker/cleanup.log

# start cron
/usr/sbin/crond -f -l 9 -L /var/log/cron.log

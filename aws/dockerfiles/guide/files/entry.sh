#!/bin/sh

# set system wide env variables, so they are available to ssh connections
/usr/bin/env > /etc/environment

echo "Initialize logging for guide daemons"
# setup symlink to output logs from relevant scripts to container logs
ln -s /proc/1/fd/1 /var/log/docker/refresh.log
ln -s /proc/1/fd/1 /var/log/docker/watcher.log
ln -s /proc/1/fd/1 /var/log/docker/cleanup.log
ln -s /proc/1/fd/1 /var/log/docker/bouncer.log
ln -s /proc/1/fd/1 /var/log/docker/vacuum.log

# start cron
/usr/sbin/crond -f -l 9 -L /var/log/cron.log

#!/bin/bash

# set system wide env variables, so they are available to ssh connections
/usr/bin/env > /etc/environment

# start cron
/usr/sbin/crond -f -l 9 -L /var/log/docker/cron.log &

echo "Initializing logs" > /var/log/docker/buoy.log
echo "Initializing logs" > /var/log/docker/refresh.log
echo "Initializing logs" > /var/log/docker/vacuum.log
echo "Initializing logs" > /var/log/docker/kick_upgrade.log
echo "Initializing logs" > /var/log/docker/kick_rejoin.log
echo "Initializing logs" > /var/log/docker/kick_nodewatch.log

echo "Initializing logs" > /var/log/docker/azupg_listener_modules.log
echo "Initializing logs" > /var/log/docker/azupg_listener_actions.log
echo "Initializing logs" > /var/log/docker/azupg_listener_daemon.log
echo "Initializing logs" > /var/log/docker/aznodewatch_modules.log
echo "Initializing logs" > /var/log/docker/aznodewatch_daemon.log
echo "Initializing logs" > /var/log/docker/aznodewatch_actions.log
echo "Initializing logs" > /var/log/docker/azrejoin_modules.log
echo "Initializing logs" > /var/log/docker/azrejoin_actions.log
echo "Initializing logs" > /var/log/docker/azrejoin_daemon.log
echo "Initializing logs" > /var/log/docker/azupgrade_modules.log
echo "Initializing logs" > /var/log/docker/azupgrade_actions.log

tail -F /var/log/docker/*

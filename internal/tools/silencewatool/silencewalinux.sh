#!/bin/bash

echo "Existing contents of /opt/waagent.conf:"
docker exec agent /bin/sh -c "cat /opt/waagent.conf"
echo "=============="
echo "Switch Logs.Verbose of /opt/waagent.conf from 'y' to 'n' ..."
echo "=============="
docker exec agent /bin/sh -c "sed -i -e s/Logs.Verbose=y/Logs.Verbose=n/g /opt/waagent.conf"
echo "New contents of /opt/waagent.conf:"
docker exec agent /bin/sh -c "cat /opt/waagent.conf"

echo "**************"
echo "Current PID of waagent:"
docker exec agent /bin/sh -c "ps -ef | grep 'root * 0:00 python /usr/sbin/waagent'"

docker exec agent /bin/sh -c "ps -ef | grep 'root * 0:00 python /usr/sbin/waagent' | sed -e 's/^[ \t]*//' | cut -d' ' -f 1" > /tmp/WAPID
echo "Restart waagent $PID..."
PID=$(cat /tmp/WAPID) 
docker exec agent /bin/sh -c "kill $PID"
sleep 2
echo "New PID of waagent:"
docker exec agent /bin/sh -c "ps -ef | grep 'root * 0:00 python /usr/sbin/waagent'"
echo "Done"

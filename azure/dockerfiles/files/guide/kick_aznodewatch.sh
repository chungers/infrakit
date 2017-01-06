#!/bin/bash

PIDFILE=/var/nodewatch.pid
if [ -f $PIDFILE ]
then
  PID=$(cat $PIDFILE)
  ps -o pid | grep $PID
  if [ $? -eq 0 ]
  then
    echo "Job is already running"
    exit 0
  else
    ## Process not found assume not running
    echo $$ > $PIDFILE
    if [ $? -ne 0 ]
    then
      echo "Could not create PID file"
      exit 1
    fi
  fi
else
  echo $$ > $PIDFILE
  if [ $? -ne 0 ]
  then
    echo "Could not create PID file"
    exit 1
  fi
fi

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')
## only run in current leader
if [[ "$IS_LEADER" == "true" ]]; then
    python -u /usr/bin/aznodewatch.py >> /var/log/nodewatch.log
fi

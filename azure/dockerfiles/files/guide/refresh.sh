#!/bin/bash
# this script refreshes the swarm tokens in azure table if they have changed.
if [ "$ROLE" == "WORKER" ] ; then
    # this doesn't run on workers, only managers.
    exit 0
fi

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
    # we are the leader, We only need to call once, so we only call from the current leader.
    MYIP=$(ifconfig eth0 | grep "inet addr:" | cut -d: -f2 | cut -d" " -f1)
    CURRENT_MANAGER_IP=$(python /usr/bin/azureleader.py get-ip)
    echo "Current manager IP = $CURRENT_MANAGER_IP ; my IP = $MYIP"

    if [ "$CURRENT_MANAGER_IP" == "$MYIP" ]; then
        echo "Swarm Manager IP changed, updating azure table with new ip"
        python /usr/bin/azureleader.py insert-ip $MYIP
    fi

fi

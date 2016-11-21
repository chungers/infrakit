#!/bin/bash
# this script refreshes the swarm tokens in azure table if they have changed.
if [ "$NODE_TYPE" == "worker" ] ; then
    # this doesn't run on workers, only managers.
    exit 0
fi

get_manager_token()
{
    if [ -n "$MANAGER_IP" ]; then
        export MANAGER_TOKEN=$(wget -qO- http://$MANAGER_IP:9024/token/manager/)
        echo "MANAGER_TOKEN=$MANAGER_TOKEN"
    else
        echo "MANAGER_TOKEN can't be found yet. MANAGER_IP isn't set yet."
    fi
}

get_worker_token()
{
    if [ -n "$MANAGER_IP" ]; then
        export WORKER_TOKEN=$(wget -qO- http://$MANAGER_IP:9024/token/worker/)
        echo "WORKER_TOKEN=$WORKER_TOKEN"
    else
        echo "WORKER_TOKEN can't be found yet. MANAGER_IP isn't set yet."
    fi
}

IS_LEADER=$(docker node inspect self -f '{{ .ManagerStatus.Leader }}')

if [[ "$IS_LEADER" == "true" ]]; then
    # we are the leader, We only need to call once, so we only call from the current leader.
    MYIP=$(ifconfig eth0 | grep "inet addr:" | cut -d: -f2 | cut -d" " -f1)
    CURRENT_MANAGER_IP=$(python /usr/bin/azuretokens.py get-ip)
    echo "Current manager IP = $CURRENT_MANAGER_IP ; my IP = $MYIP"

    if [ "$CURRENT_MANAGER_IP" == "$MYIP" ]; then
        echo "Swarm Manager IP changed, updating azure table with new ip"
        python /usr/bin/azuretokens.py insert-ip $MYIP
    fi

fi

#!/bin/bash
AZ_GUIDE_CONTAINER=editions_guide
echo "Copying upgrade script ..."
docker cp /opt/azupgrade.py $AZ_GUIDE_CONTAINER:/usr/bin/azupgrade.py
docker cp /opt/azupgrade_log_cfg.json $AZ_GUIDE_CONTAINER:/etc/azupgrade_log_cfg.json
docker cp /opt/azendpt.py $AZ_GUIDE_CONTAINER:/usr/bin/azendpt.py
docker cp /opt/aztags.py $AZ_GUIDE_CONTAINER:/usr/bin/aztags.py
docker cp /opt/dockerutils.py $AZ_GUIDE_CONTAINER:/usr/bin/dockerutils.py
# for some reason the line: DOCKER_CHANNEL=$(docker exec $AZ_GUIDE_CONTAINER /usr/bin/aztags.py channel)
# resulted in error exec: editions_guide: not found
# so as a workaround dump channel to a file and read from that.
docker exec $AZ_GUIDE_CONTAINER /usr/bin/aztags.py channel > /tmp/channel
DOCKER_CHANNEL=$(cat /tmp/channel)
echo "Kicking off upgrade to https://download.docker.com/azure/$DOCKER_CHANNEL/$DOCKER_VERSION/Docker.tmpl ..."
docker exec -ti $AZ_GUIDE_CONTAINER /usr/bin/azupgrade.py https://download.docker.com/azure/$DOCKER_CHANNEL/$DOCKER_VERSION/Docker.tmpl

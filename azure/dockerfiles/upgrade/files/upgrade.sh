#!/bin/bash
AZ_GUIDE_CONTAINER=editions_guide
CUSTOM_DATA_FILE=/var/lib/waagent/CustomData

copy_file_to_guide()
{
    docker cp $1 $AZ_GUIDE_CONTAINER:$2
}

parse_export_value()
{
    expval=$(grep $1= $CUSTOM_DATA_FILE | sed -e 's/export .[A-Z|_]*\=\"\(.*\)\"/\1/')
}

echo "Copying upgrade scripts ..."
copy_file_to_guide /opt/azupgrade.py /usr/bin/azupgrade.py
copy_file_to_guide /opt/azupgrade_log_cfg.json /etc/azupgrade_log_cfg.json
copy_file_to_guide /opt/azendpt.py /usr/bin/azendpt.py
copy_file_to_guide /opt/aztags.py /usr/bin/aztags.py
copy_file_to_guide /opt/dockerutils.py /usr/bin/dockerutils.py

DTR_HUB_TAG=$(docker images --filter=reference='docker/dtr-*' --format "{{.Tag}}" | head -n 1)
DOCKER_EE=$(docker -v | grep "\-ee")

if [ -z "$DTR_HUB_TAG" ] && [ -z "$DOCKER_EE" ]; then
    # Docker CE
    docker exec $AZ_GUIDE_CONTAINER /usr/bin/aztags.py channel > /tmp/channel
    DOCKER_CHANNEL=$(cat /tmp/channel)
else
    #Docker EE or DDC
    DOCKER_CHANNEL=$(echo "${DOCKER_VERSION%.*}")
fi

TEMPLATE_URL_PREFIX=https://download.docker.com/azure/$DOCKER_CHANNEL/$DOCKER_VERSION

# presence of DTR images (and associated tag) indicate a DDC setup
if [ -z "$DTR_HUB_TAG" ]; then
    # for EE and CE we don't need to parse out additional parameters
    TEMPLATE_URL=$TEMPLATE_URL_PREFIX/Docker.tmpl
    echo "Kicking off upgrade to $TEMPLATE_URL ..."
    docker exec -ti $AZ_GUIDE_CONTAINER /usr/bin/azupgrade.py $TEMPLATE_URL
else
    # for DDC we need additional parameters
    parse_export_value UCP_ADMIN_PASSWORD
    UCP_ADMIN_PASSWORD=$expval

    parse_export_value UCP_ADMIN_USER
    UCP_ADMIN_USER=$expval

    parse_export_value UCP_ELB_HOSTNAME
    UCP_ELB_HOSTNAME=$expval

    parse_export_value UCP_LICENSE
    UCP_LICENSE=$(echo -n  "$expval" | base64 -d)

    TEMPLATE_URL=$TEMPLATE_URL_PREFIX/Docker-DDC.tmpl
    echo "Kicking off upgrade to $TEMPLATE_URL ..."
    docker exec -ti -e DTR_HUB_TAG=$DTR_HUB_TAG -e UCP_ADMIN_USER=$UCP_ADMIN_USER -e UCP_ADMIN_PASSWORD=$UCP_ADMIN_PASSWORD -e UCP_ELB_HOSTNAME=$UCP_ELB_HOSTNAME -e UCP_LICENSE=$UCP_LICENSE $AZ_GUIDE_CONTAINER /usr/bin/azupgrade.py $TEMPLATE_URL
fi

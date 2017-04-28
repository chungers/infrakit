if [ $ENABLE_EFS == 'yes' ] ; then
    docker plugin install --alias cloudstor:aws --grant-all-permissions docker4x/cloudstor:$DOCKER_FOR_IAAS_VERSION CLOUD_PLATFORM=AWS EFS_ID_REGULAR=$EFS_ID_REGULAR EFS_ID_MAXIO=$EFS_ID_MAXIO DEBUG=1
fi

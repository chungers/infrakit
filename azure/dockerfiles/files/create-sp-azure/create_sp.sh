#!/bin/bash

usage() {
    echo "usage: ${PROG} [-e Public] app_name [resource_group_name resource_group_location]"
    echo "       ${PROG} -e Gov app_name user_name [resource_group_name resource_group_location]"
    echo "    app_name: Name of the new Azure AD Application used for authentication"
    echo "    user_name: Username for the Azure account in Azure government regions"
    echo "    resource_group_name: Name of the new Azure Resource Group where SP will be scoped to"
    echo "    resource_group_location: Location of the new Azure Resource Group"
}

MAX_RETRIES=20
PROG="${PROG:-$0}"
ENV_FLAG=$1
# default env is public - maintain compat with previous docs/usage
ENV="Public"

if [[ $ENV_FLAG == "-e" ]]; then
    shift
    ENV=$1
    shift
fi

if [[ $ENV == "Gov" ]]; then
    if [ $# -ne 2 ] && [ $# -ne 4 ]; then
        usage
        exit 1
    fi
    APP_NAME=$1
    USER_NAME=$2
    RG_NAME=$3
    RG_LOC=$4

    azure login --username $USER_NAME --environment "AzureUSGovernment"

elif [[ $ENV == "Public" ]]; then
    if [ $# -ne 1 ] && [ $# -ne 3 ]; then
        usage
        exit 1
    fi
    APP_NAME=$1
    RG_NAME=$2
    RG_LOC=$3

    azure login

else
    usage
    exit 1
fi

if [ -z "$SUBSCRIPTION_ID" ]; then
    echo "The following subscriptions were retrieved from your Azure account"
    PS3='Please select the subscription option number to use for Docker swarm resources: '
    options=($(azure account list --json | jq -r 'map(select(.state == "Enabled"))|.[]|.id + ":" + .name' | sed -e 's/ /_/g'))
    select opt in "${options[@]}"
    do
        SUBSCRIPTION_ID=`echo $opt | awk -F ':' '{print $1}'`
        break
    done
fi

echo "Using subscription ${SUBSCRIPTION_ID}"

azure account set ${SUBSCRIPTION_ID}

TENANT_ID=$(azure account list --json | jq "map(select(.isDefault == true)) | .[0].tenantId" | sed -e 's/\"//g')

if [[ "" == ${TENANT_ID} ]]; then
    echo "Not logged in.  Cannot determine tenant id."
    exit 1
fi


echo "Creating AD application ${APP_NAME}"

PASSWORD=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

mkdir -p /var/lib/azure
echo ${PASSWORD} > /var/lib/azure/passwd
chmod 600 /var/lib/azure/passwd

azure ad app create --name ${APP_NAME} \
      --home-page https://${APP_NAME} \
      --identifier-uris https://${APP_NAME} \
      --password ${PASSWORD} \
      --json > /var/lib/azure/ad_app_create.json

APP_ID=$(jq .appId /var/lib/azure/ad_app_create.json | sed -e s/\"//g)

echo "Created AD application, APP_ID=${APP_ID}"
if [[ "" == ${APP_ID} ]]; then
    echo "Cannot create application or determine application id."
    exit 1
fi

echo "Creating AD App ServicePrincipal"
azure ad sp create --applicationId ${APP_ID}  --json > /var/lib/azure/ad_sp_create.json

SP_OBJECT_ID=$(jq .objectId  /var/lib/azure/ad_sp_create.json | sed -e 's/\"//g')

echo "Created ServicePrincipal ID=${SP_OBJECT_ID}"
if [[ "" == ${SP_OBJECT_ID} ]]; then
    echo "Cannot create service principal or determine its object id."
    exit 1
fi

if [ ! -z ${RG_NAME} ]; then
    echo "Create new Azure Resource Group "${RG_NAME}" in "${RG_LOC}""
    azure group create "${RG_NAME}" "${RG_LOC}"
    if [[ $? -ne 0 ]]; then
        echo "Resource Group creation failed. Details:"
        cat /root/.azure/azure.err
        exit 1
    fi
    echo "Resource Group "${RG_NAME}" created "
fi

echo "Waiting for account updates to complete before proceeding ..."

sleep 30
retries=0
status=1
while [ ${status} -ne 0 ] && [ ${retries} -lt ${MAX_RETRIES} ];
do

    if [ -z ${RG_NAME} ]; then
        echo "Creating role assignment for ${SP_OBJECT_ID} for subscription ${SUBSCRIPTION_ID}"
        azure role assignment create --objectId ${SP_OBJECT_ID} --roleName Contributor \
            --scope /subscriptions/${SUBSCRIPTION_ID}/
        status=$?
    else
        echo "Creating role assignment for ${SP_OBJECT_ID} scoped to ${RG_NAME}"
        azure role assignment create --objectId ${SP_OBJECT_ID} --roleName Contributor \
            --resource-group "${RG_NAME}"
        status=$?
    fi

    if [ ${status} -ne 0 ]; then
        echo "Details from last failure:"
        cat /root/.azure/azure.err
        echo "Wait before retrying ..."
        sleep 30
    fi
    retries=$((retries+1))
done

if [ ${status} -ne 0 ]; then
    echo "Role assignment creation failed for Azure."
    echo "we generated the following:"
    echo "AD ServicePrincipal App ID:       ${APP_ID}"
    echo "AD ServicePrincipal App Secret:   ${PASSWORD}"
    exit 1
fi
echo "Successfully created role assignment for ${SP_OBJECT_ID}"

echo "Test login..."
echo "Waiting for roles to take effect ..."
sleep 30
retires=0
status=1
while [ ${status} -ne 0 ] && [ ${retries} -lt ${MAX_RETRIES} ];
do
    azure login --service-principal --tenant ${TENANT_ID} --username ${APP_ID} --password ${PASSWORD}
    status=$?

    if [[ ${status} -ne 0 ]]; then
        echo "Details from last failure:"
        cat /root/.azure/azure.err
        echo "Wait before retrying ..."
        sleep 30
    fi
    retries=$((retries+1))
done

if [ ${status} -ne 0 ]; then
    echo "Login failed for Azure with generated SP"
    echo "we generated the following:"
    echo "AD ServicePrincipal App ID:       ${APP_ID}"
    echo "AD ServicePrincipal App Secret:   ${PASSWORD}"
    exit 1
fi

echo
echo
echo

echo "Your access credentials =================================================="
echo "AD ServicePrincipal App ID:       ${APP_ID}"
echo "AD ServicePrincipal App Secret:   ${PASSWORD}"
echo "AD ServicePrincipal Tenant ID:    ${TENANT_ID}"
if [[ "" != ${RG_NAME} ]]; then
    echo "Resource Group Name:              "${RG_NAME}""
    echo "Resource Group Location:          "${RG_LOC}""
fi

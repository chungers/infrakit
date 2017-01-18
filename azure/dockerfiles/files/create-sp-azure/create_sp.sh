#!/bin/bash

PROG="${PROG:-$0}"
APP_NAME=$1

if [[ "" == ${APP_NAME} ]]; then
  echo "Must pass APP_NAME as parameter, i.e. ${PROG} [APP_NAME]."
  exit 1
fi

azure login

if [ -z "$SUBSCRIPTION_ID" ]; then
  echo "The following subscriptions were retrieved from your account"
  PS3='Please select the subscription to use: '
  options=($(azure account list --json | jq -r 'map(select(.state == "Enabled"))|.[]|.id + ":" + .name' | sed -e 's/ /_/g'))
  select opt in "${options[@]}"
  do
          SUBSCRIPTION_ID=`echo $opt | awk -F ':' '{print $1}'`
          break
  done
fi

echo "Using subscription ${SUBSCRIPTION_ID}"

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

echo "Creating role assignment for ${SP_OBJECT_ID} for subscription ${SUBSCRIPTION_ID}"

echo "Waiting for account updates to complete before proceeding ..."
sleep 30
azure role assignment create --objectId ${SP_OBJECT_ID} --roleName Contributor \
      --scope /subscriptions/${SUBSCRIPTION_ID}/ --json > /var/lib/azure/role_assignment.json

status=$?
while [[ ${status} != 0 ]];
do
    echo "Ignore previous failure and retry ..."
    sleep 30
    azure role assignment create --objectId ${SP_OBJECT_ID} --roleName Contributor \
          --scope /subscriptions/${SUBSCRIPTION_ID}/ --json > /var/lib/azure/role_assignment.json
    status=$?
done

echo "Successfully created role assignment for ${SP_OBJECT_ID} for subscription ${SUBSCRIPTION_ID}"

echo "Test login..."
azure login --service-principal --tenant ${TENANT_ID} --username ${APP_ID} --password ${PASSWORD} --json

echo 
echo
echo

echo "Your access credentials =================================================="
echo "AD ServicePrincipal App ID:       ${APP_ID}"
echo "AD ServicePrincipal App Secret:   ${PASSWORD}"
echo "AD ServicePrincipal Tenant ID:    ${TENANT_ID}"

#!/bin/bash

set -e

BASEDIR=$(dirname "$0")
BUILD_NUMBER="${BUILD_NUMBER:-0}"
STACK="ci-docker4gcp-${BUILD_NUMBER}"

export CLOUDSDK_CORE_PROJECT="${CLOUDSDK_CORE_PROJECT:-docker-for-gcp-ci}"
export CLOUDSDK_COMPUTE_ZONE="${CLOUDSDK_COMPUTE_ZONE:-us-central1-f}"

auth_gcloud() {
  echo Authenticate GCloud

  echo ${GCLOUD_SERVICE_KEY} | base64 --decode > /.config/key.json
  gcloud auth activate-service-account --key-file=/.config/key.json
}

download_templates() {
  echo Downloading templates for build ${BUILD_NUMBER}

  mkdir /templates
  gsutil -m cp -r gs://docker-for-gcp-builds/${BUILD_NUMBER}/templates/* /templates/
}

create_swarm() {
  echo Create Swarm ${STACK}

  gcloud deployment-manager deployments create ${STACK} \
    --config /templates/Docker.jinja \
    --properties managerCount:3,workerCount:1,zone:${CLOUDSDK_COMPUTE_ZONE},demoMode:true
}

check_instances_created() {
  echo Check that the instances are there

  for i in $(seq 1 120); do
    COUNT=$(gcloud compute instances list --filter="status:RUNNING AND networkInterfaces[0].network:${STACK}-network" --uri | wc -w | tr -d '[:space:]')
    echo "- ${COUNT} instances were created"

    if [ ${COUNT} -gt 5 ]; then
      echo "- ERROR: that's too many!"
      exit 1
    fi

    if [ ${COUNT} -eq 5 ]; then
      return 0
    fi

    sleep 1
  done

  echo "- ERROR: not all instances were created"
  return 1
}

check_web_service() {
  echo Check that the exposed service is reacheable

  EXTERNAL_IP=$(gcloud deployment-manager deployments describe ${STACK} --format=json | jq -r '.outputs[] | select(.name=="externalIp") | .finalValue')
  echo External IP is ${EXTERNAL_IP}

  echo Wait for the service to be up
  for i in $(seq 1 50); do
    curl -s http://${EXTERNAL_IP}:8080/ >/dev/null && break || sleep 1
  done

  echo Make sure the service is available on all nodes
  for i in $(seq 1 100); do
    curl -s http://${EXTERNAL_IP}:8080/ | grep 'Docker Demo' >/dev/null && echo "Service is available"
  done
}

update_instances() {
  echo Update the Swarm ${STACK}

  gcloud deployment-manager deployments update ${STACK} \
    --config /templates/Docker.jinja \
    --properties managerCount:3,workerCount:1,zone:${CLOUDSDK_COMPUTE_ZONE},demoMode:true,managerMachineType:n1-standard-1,workerMachineType:n1-standard-1
}

check_instances_updated() {
  echo Check that the instances are updated

  for i in $(seq 1 120); do
    COUNT=$(gcloud compute instances list --filter="status:RUNNING AND machineType=n1-standard-1 AND networkInterfaces[0].network:${STACK}-network" --uri | wc -w | tr -d '[:space:]')
    echo "- ${COUNT} instances were updated"

    if [ ${COUNT} -gt 4 ]; then
      echo "- ERROR: that's too many!"
      exit 1
    fi

    if [ ${COUNT} -eq 4 ]; then
      return 0
    fi

    sleep 1
  done

  echo "- ERROR: not all instances were updated"
  return 1
}

destroy_swarm() {
  echo Delete Swarm ${STACK}

  set +e
  INSTANCES=$(gcloud compute instances list --filter="networkInterfaces[0].network:${STACK}-network" --uri)
  set -e
  [ -n "${INSTANCES}" ] && gcloud compute instances delete -q --delete-disks=boot ${INSTANCES}

  set +e
  gcloud deployment-manager deployments describe ${STACK} >/dev/null 2>&1
  EXISTS=$?
  set -e

  if [ $EXISTS -eq 0 ]; then
    gcloud --verbosity=none deployment-manager deployments describe ${STACK} && gcloud deployment-manager deployments delete -q ${STACK} || true
  fi

  set +e
  DISKS=$(gcloud compute disks list --filter="name~${STACK}-manager-" --uri)
  set +e
  [ -n "${DISKS}" ] && gcloud compute disks delete -q ${DISKS}
}

check_instances_gone() {
  echo Check that the instances are gone

  COUNT=$(gcloud compute instances list --filter="networkInterfaces[0].network:${STACK}-network" --uri | wc -w | tr -d '[:space:]')

  if [ ${COUNT} -eq 0 ]; then
    echo "- All instances are gone"
    return
  fi

  echo "ERROR: ${COUNT} instances are still around"
  exit 1
}

if [ -z "${GCLOUD_SERVICE_KEY}" ]; then
  echo "Needs GCLOUD_SERVICE_KEY env variable"
  exit 1
fi
auth_gcloud
download_templates
destroy_swarm
create_swarm
check_instances_created
check_web_service
update_instances
check_instances_updated
check_web_service
destroy_swarm
check_instances_gone

exit 0

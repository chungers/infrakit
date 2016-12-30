#!/bin/bash

set -e

BASEDIR=$(dirname "$0")
GCLOUD="docker run -e CLOUDSDK_CORE_PROJECT --rm -v gcloud-config:/.config -v $(pwd)/${BASEDIR}/../templates:/templates google/cloud-sdk gcloud"
STACK="ci-docker4gcp-${CIRCLE_BUILD_NUM:-local}"

export CLOUDSDK_CORE_PROJECT="docker4x"
export CLOUDSDK_COMPUTE_ZONE="us-central1-f"

cleanup() {
  echo Clean up

  docker volume rm gcloud-config 2>/dev/null || true
}

auth_gcloud() {
  echo Authenticate GCloud

  docker volume create --name gcloud-config
  docker run --rm -e GCLOUD_SERVICE_KEY -v gcloud-config:/.config google/cloud-sdk bash -c 'echo ${GCLOUD_SERVICE_KEY} | base64 --decode > /.config/key.json'
  docker run --rm -v gcloud-config:/.config google/cloud-sdk gcloud auth activate-service-account --key-file=/.config/key.json
}

create_swarm() {
  echo Create Swarm ${STACK}

  ${GCLOUD} deployment-manager deployments create ${STACK} \
    --config /templates/swarm.jinja \
    --properties managerCount=3,workerCount=1,zone=${CLOUDSDK_COMPUTE_ZONE}
}

check_instances_created() {
  echo Check that the instances are there

  for i in $(seq 1 120); do
    COUNT=$(${GCLOUD} compute instances list --filter="status:RUNNING AND networkInterfaces[0].network:${STACK}-network" --uri | wc -w | tr -d '[:space:]')
    echo "- ${COUNT} instances where created"

    if [ ${COUNT} -gt 4 ]; then
      echo "- ERROR: that's too many!"
      exit 1
    fi

    if [ ${COUNT} -eq 4 ]; then
      return
    fi

    sleep 1
  done
}

destroy_swarm() {
  echo Delete Swarm ${STACK}

  INSTANCES=$(${GCLOUD} compute instances list --filter="networkInterfaces[0].network:${STACK}-network" --uri)
  [ -n "${INSTANCES}" ] && ${GCLOUD} compute instances delete -q --delete-disks=boot ${INSTANCES}

  set +e
  ${GCLOUD} deployment-manager deployments describe ${STACK} >/dev/null 2>&1
  EXISTS=$?
  set -e

  if [ $EXISTS -eq 0 ]; then
    ${GCLOUD} --verbosity=none deployment-manager deployments describe ${STACK} && ${GCLOUD} deployment-manager deployments delete -q ${STACK} || true
  fi
}

check_instances_gone() {
  echo Check that the instances are gone

  COUNT=$(${GCLOUD} compute instances list --filter="networkInterfaces[0].network:${STACK}-network" --uri | wc -w | tr -d '[:space:]')

  if [ ${COUNT} -eq 0 ]; then
    echo "- All instances are gone"
    return
  fi

  echo "ERROR: ${COUNT} instances are still around"
  exit 1
}

[ -n "${GCLOUD_SERVICE_KEY}" ] || exit 0
[ -n "${CI}" ] || cleanup
auth_gcloud
[ -n "${CI}" ] || destroy_swarm
create_swarm
check_instances_created
destroy_swarm
check_instances_gone

exit 0

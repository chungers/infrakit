#!/bin/sh
#
# This script is meant for quick & easy install via:
#   'curl -sSL http://get.docker-gcp.com/ | sh'
# or:
#   'wget -qO- http://get.docker-gcp.com/ | sh'
#
set -e

unset DOCKER_HOST

echome() {
  echo "$@"
  $@
}

metadata() {
    curl -s "http://metadata.google.internal/computeMetadata/v1/$1" \
        -H "Metadata-Flavor: Google"
}

getval() {
    PROJECT=$(gcloud config list project --format=json 2>/dev/null | jq -r .core.project)
    AUTH=$(metadata instance/service-accounts/default/token | jq -r ".access_token")

    curl -s "https://runtimeconfig.googleapis.com/v1beta1/projects/${PROJECT}/configs/swarm-config/variables/$1" \
        -H "Authorization: Bearer ${AUTH}" | jq -r ".text // empty"
}

echo
echo "Let's install Docker!"
echo "First, let's answer a few questions:"
echo
echo -n "How many managers? (3, 5 or 7. Default is 3) "
read managerCount </dev/tty
case ${managerCount} in
  "")     managerCount=3;;
  3|5|7)  ;;
  *)      echo "Invalid value"; exit 1;;
esac

echo
echo -n "How many workers? (Default is 1) "
read workerCount </dev/tty
case ${workerCount} in
  0)      echo "There must be at least one worker"; exit 1;;
  ""|1)   workerCount=1;;
  [0-9]*) ;;
  *)      echo "Invalid value"; exit 1;;
esac

echo
echo "If you don't want to curl this script, you can run the following command directly:"
echome gcloud deployment-manager deployments create docker \
  --config https://storage.googleapis.com/docker-template/swarm.py \
  --properties managerCount=${managerCount},workerCount=${workerCount}

echo
echo "Build a container to open an ssh tunnel..."
/usr/bin/docker ps -aq -f "name=tunnel" | xargs --no-run-if-empty /usr/bin/docker rm -f
/usr/bin/docker build -q -t tunnel - >/dev/null 2>&1 << EOF
FROM alpine:3.4
RUN apk --update add openssh && rm -Rf /var/lib/cache/apk/*
ENTRYPOINT ["/usr/bin/ssh", "-i", "~/.ssh/google_compute_engine", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null", "-o", "CheckHostIP=no", "-NL", "0.0.0.0:2374:/var/run/docker.sock"]
EOF

echo
echo -n "Wait for the leader..."
while [ -z "$(getval zone)" ]; do
    echo -n "."
    sleep 2
done
echo

echo
echo "Retrieve Swarm properties..."
LEADER_IP=$(getval leader-ip)
LEADER_NAME=$(getval leader-name)
ZONE_URI=$(getval zone)
ZONE=$(echo ${ZONE_URI} | awk -F/ '{print $NF}')
REGION=$(echo ${ZONE} | awk -F- '{print $1 "-" $2}')
LEADER_NAT_IP=$(gcloud compute instances describe --zone ${ZONE} ${LEADER_NAME} --format="value(networkInterfaces[0].accessConfigs[0].natIP)")
EXTERNAL_IP=$(gcloud compute addresses describe --region ${REGION} docker-ip --format=json | jq -r '.address')

gcloud compute ssh --zone ${ZONE} ${LEADER_NAME} --command "sudo usermod -aG docker $(whoami)" >/dev/null 2>&1

cat > ~/README-DOCKER << EOF
                        ##         .
                   ## ## ##        ==
                ## ## ## ## ##    ===
            /"""""""""""""""""\___/ ===
       ~~~ {~~ ~~~~ ~~~ ~~~~ ~~~ ~ /  ===- ~~~
            \______ o           __/
              \    \         __/
               \____\_______/

Welcome to Docker!

You can ssh into the Swarm with:
  gcloud compute ssh --zone ${ZONE} ${LEADER_NAME}

Or connect via an ssh tunnel with:
  /usr/bin/docker run -d --name tunnel -v \$HOME/.ssh:/root/.ssh -p 2374:2374 tunnel \$(whoami)@${LEADER_NAT_IP}
  export DOCKER_HOST=localhost:2374
  docker ps

The services are published on:
  ${EXTERNAL_IP}

To uninstall Docker, run this command:
  gcloud deployment-manager deployments delete docker

Have fun!
EOF

echo
cat README-DOCKER

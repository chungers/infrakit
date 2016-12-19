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

echo
echo "Let's install Docker!"
echo "First, let's answer a few questions:"
echo
echo -n "How many managers? (1, 3 or 5. Default is 3) "
read managerCount </dev/tty
case ${managerCount} in
  "")    managerCount=3;;
  1|3|5) ;;
  *)     echo "Invalid value"; exit 1;;
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
echo "Retrieve Swarm properties..."
DEPLOYMENT=$(gcloud deployment-manager deployments describe docker --format='table(outputs)')
ZONE=$(echo "${DEPLOYMENT}" | awk '/zone/ {print $2}')
LEADER_IP=$(echo "${DEPLOYMENT}" | awk '/leaderIp/ {print $2}')
EXTERNAL_IP=$(echo "${DEPLOYMENT}" | awk '/externalIp/ {print $2}')

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
  gcloud compute ssh --zone ${ZONE} manager-1

Or connect via an ssh tunnel with:
  /usr/bin/docker run -d --name tunnel -v \$HOME/.ssh:/root/.ssh -p 2374:2374 tunnel \$(whoami)@${LEADER_IP}
  export DOCKER_HOST=localhost:2374
  docker ps

The services are published on:
  ${EXTERNAL_IP}

To uninstall Docker, run these commands:
  gcloud compute instances delete \$(gcloud compute instances list --filter='tags.items ~ swarm' --uri)
  gcloud deployment-manager deployments delete docker

Have fun!
EOF

echo
cat README-DOCKER

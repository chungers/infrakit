#!/bin/sh
#
# This script is meant for quick & easy install via:
#   'curl -sSL http://get.docker-gcp.com/ | sh'
# or:
#   'wget -qO- http://get.docker-gcp.com/ | sh'
#
set -e

echome() {
  echo "$@"
  $@
}

echo "Let's deploy Docker!"
echo "First, here's a few questions:"
echo
echo "How many managers? (3,5 or 7. Default is 3)"
read managerCount </dev/tty
case ${managerCount} in
  "")     managerCount=3;;
  3|5|7)  ;;
  *)      echo "Invalid value"; exit 1;;
esac

echo
echo "How many workers? (Default is 1)"
read workerCount </dev/tty
case ${workerCount} in
  0)      echo "There must be at least one worker"; exit 1;;
  ""|1)   workerCount=1;;
  [0-9]*) ;;
  *)      echo "Invalid value"; exit 1;;
esac

echome gcloud deployment-manager deployments create docker \
  --config https://storage.googleapis.com/docker-template/swarm.py \
  --properties managerCount=${managerCount},workerCount=${workerCount}

cat << EOF
Welcome to Docker

                        ##         .
                   ## ## ##        ==
                ## ## ## ## ##    ===
            /"""""""""""""""""\___/ ===
       ~~~ {~~ ~~~~ ~~~ ~~~~ ~~~ ~ /  ===- ~~~
            \______ o           __/
              \    \         __/
               \____\_______/

To delete the Swarm, run this command:
   gcloud deployment-manager deployments delete docker

Have fun!
EOF

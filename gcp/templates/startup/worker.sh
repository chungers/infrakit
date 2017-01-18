#!/bin/bash

set -ex

if [ "$(docker info -f '{{ "{{" }}.Swarm.LocalNodeState{{ "}}" }}')" != "inactive" ]; then
  echo "Skipping the startup script"
  exit 0
fi

DOCKER_FOR_IAAS_VERSION='{{ VERSION }}'

docker_run='docker run --label com.docker.editions.system --log-driver=json-file'
docker_daemon="$docker_run -d --restart=always"
docker_client='-v /var/run/docker.sock:/var/run/docker.sock'

echo Start sshd

shell_image="docker4x/shell-gcp:$DOCKER_FOR_IAAS_VERSION"
for i in $(seq 1 60); do docker pull $shell_image && break || sleep 1; done

$docker_run -it --name=etc \
  --user root \
  -v /etc \
  --entrypoint ssh-keygen \
  $shell_image \
  -A

$docker_daemon --name=accounts \
  -v /dev/log:/dev/log \
  -v /home:/home \
  --volumes-from=etc \
  $shell_image \
  /usr/bin/google_accounts_daemon -d

$docker_daemon --name=shell -p 22:22 \
  $docker_client \
  -v /usr/bin/docker:/usr/bin/docker \
  -v /var/log:/var/log \
  -v /home:/home \
  --volumes-from=etc \
  $shell_image

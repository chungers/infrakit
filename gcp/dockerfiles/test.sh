#!/bin/bash

docker volume create --name sshkey

docker run --log-driver=json-file -ti --rm \
  --user root \
  -v sshkey:/etc/ssh \
  --entrypoint ssh-keygen \
  docker4x/shell-gcp \
  -A

docker run --log-driver=json-file --name=shell --restart=always -d -p 22:22 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /usr/bin/docker:/usr/bin/docker \
  -v /dev/log:/dev/log \
  -v /var/log:/var/log \
  -v /home:/home \
  -v sshkey:/etc/ssh \
  docker4x/shell-gcp

docker exec shell /usr/bin/google_accounts_daemon -d

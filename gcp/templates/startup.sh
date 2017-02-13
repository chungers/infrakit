#!/bin/sh

set -ex

echo This is a {{type}} node

shell_image="docker4x/shell-gcp:{{ VERSION }}"
guide_image="docker4x/guide-gcp:{{ VERSION }}"
lb_image="docker4x/l4controller-gcp:{{ VERSION }}"
infrakit_image="docker4x/infrakit-gcp:{{ VERSION }}"

docker_run='docker run --label com.docker.editions.system --log-driver=json-file'
docker_daemon="$docker_run --rm -d"
docker_socket='-v /var/run/docker.sock:/var/run/docker.sock'
docker_cli='-v /usr/bin/docker:/usr/bin/docker'

function dockerPull {
  for i in $(seq 1 60); do docker pull $1 && break || sleep 1; done
}

{% if (type in ['leader']) %}
echo Initialize Swarm

docker node inspect self || docker swarm init --advertise-addr eth0:2377 --listen-addr eth0:2377
docker node inspect self
# send identify message
/usr/bin/buoy -event=identify -iaas_provider=gcp
{% endif -%}

{% if (type in ['manager', 'leader']) %}
echo Start infrakit

dockerPull ${infrakit_image}
$docker_daemon --name=infrakit $docker_socket $docker_cli $infrakit_image
{% endif -%}

echo Start sshd

dockerPull ${shell_image}

docker inspect etc >/dev/null 2>&1 || $docker_run --name=etc -v /etc $shell_image true
$docker_run --volumes-from=etc $shell_image /usr/bin/ssh-keygen.sh

$docker_daemon --name=accounts \
  -v /dev/log:/dev/log \
  -v /home:/home \
  --volumes-from=etc \
  $shell_image \
  /usr/bin/google_accounts_daemon

$docker_daemon --name=ipforwarding \
  -v /dev/log:/dev/log \
  --cap-add=NET_ADMIN \
  --net=host \
  $shell_image \
  /usr/bin/google_ip_forwarding_daemon -d

$docker_daemon --name=shell \
  -p 22:22 \
  $docker_socket \
  $docker_cli \
  -v /var/log:/var/log \
  -v /home:/home \
  --volumes-from=etc \
  --net=host \
  $shell_image

echo Start guide

dockerPull ${guide_image}
$docker_daemon --name=guide \
  -e RUN_VACUUM="{{ properties['enableSystemPrune'] }}" \
  $docker_socket \
  $docker_cli \
  $guide_image

{% if (type in ['manager', 'leader']) %}
echo Start Load Balancer Listener

dockerPull ${lb_image}
$docker_daemon --name=lbcontroller $docker_socket $lb_image run --log=5
{% endif -%}

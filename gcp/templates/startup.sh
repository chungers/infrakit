#!/bin/sh

set -ex

echo This is a {{type}} node

export DOCKER_FOR_IAAS_VERSION="{{ EDITIONS_VERSION }}"
export ACCOUNT_ID="$(curl -sH 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/email)"
export REGION="{{ REGION }}"
export CHANNEL="edge"
export NODE_TYPE="{{ type }}"

shell_image="docker4x/shell-gcp:{{ EDITIONS_VERSION }}"
guide_image="docker4x/guide-gcp:{{ EDITIONS_VERSION }}"
lb_image="docker4x/lb-controller-gcp:{{ EDITIONS_VERSION }}"
infrakit_image="docker4x/infrakit-gcp:{{ EDITIONS_VERSION }}"
cloudstor_image="docker4x/cloudstor-gcp:{{ EDITIONS_VERSION }}"

docker_run='docker container run --label com.docker.editions.system --log-driver=json-file'
docker_daemon="$docker_run --rm -d"
docker_daemon_restart="$docker_run -d"
docker_socket='-v /var/run/docker.sock:/var/run/docker.sock'
docker_cli='-v /usr/bin/docker:/usr/bin/docker'
# alpine installs in /usr/bin - debian in /usr/local/bin
gplatform_dir=/usr/local/bin

{% if (type in ['leader']) %}
echo Initialize Swarm

cat << EOF > /etc/docker/daemon.json
{
  "labels": [
  "infrakit-link=bootstrap",
]
}
EOF
kill -s HUP $(cat /var/run/docker.pid)

docker node inspect self >/dev/null 2>&1 || docker swarm init --advertise-addr eth0:2377 --listen-addr eth0:2377
docker node inspect self

$docker_run --rm \
  -e NODE_TYPE \
  -e ACCOUNT_ID \
  -e REGION \
  $docker_socket \
  $docker_cli \
  $guide_image \
  /usr/bin/buoy.sh "identify"

$docker_run --rm \
  -e NODE_TYPE \
  -e DOCKER_FOR_IAAS_VERSION \
  -e ACCOUNT_ID \
  -e REGION \
  -e CHANNEL \
  $docker_socket \
  $docker_cli \
  $guide_image \
  /usr/bin/buoy.sh "swarm:init"

{% endif -%}

{% if (type in ['manager', 'leader']) %}
echo Start infrakit

$docker_daemon --name=infrakit $docker_socket $docker_cli $infrakit_image
{% endif -%}

echo Start sshd

docker container inspect etc >/dev/null 2>&1 || $docker_run --name=etc -v /etc $shell_image true
$docker_run --volumes-from=etc $shell_image /usr/bin/ssh-keygen.sh

$docker_daemon_restart --name=accounts \
  -v /dev/log:/dev/log \
  -v /home:/home \
  --volumes-from=etc \
  $shell_image \
  $gplatform_dir/google_accounts_daemon

$docker_daemon --name=ipforwarding \
  -v /dev/log:/dev/log \
  --cap-add=NET_ADMIN \
  --net=host \
  $shell_image \
  $gplatform_dir/google_ip_forwarding_daemon -d

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

$docker_daemon --name=guide \
  -e NODE_TYPE \
  -e RUN_VACUUM="{{ properties['enableSystemPrune'] }}" \
  -e DOCKER_FOR_IAAS_VERSION \
  -e ACCOUNT_ID \
  -e REGION \
  -e CHANNEL \
  $docker_socket \
  $docker_cli \
  $guide_image

echo Install cloudstor-gcp
docker plugin install --alias cloudstor:gcp --grant-all-permissions $cloudstor_image CLOUD_PLATFORM=GCP

{% if (type in ['manager', 'leader']) %}
echo Start Load Balancer Listener

$docker_daemon --name=lbcontroller $docker_socket $lb_image run --log=5
{% endif -%}

{% if ((type in ['leader']) and (properties['demoMode'])) %}
docker service create --name demo -p 8080:8080 ehazlett/docker-demo
{% endif -%}

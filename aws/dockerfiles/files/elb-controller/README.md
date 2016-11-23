# ELB controller
Load balancer code for docker for AWS

## Updating docker client vendoring package:
At the top level directory of the repo, run this:

```
$ govendor fetch github.com/docker/docker/client
```
this will fetch the latest docker/docker client package and updates the vendor.json

## Building images

Inside of the `aws/dockerfiles/elb-controller/container` directory.

```
VERSION=aws-v1.13.0-rc2-beta12
DOCKER_TAG=$VERSION DOCKER_PUSH=true DOCKER_TAG_LATEST=false make -k container
```

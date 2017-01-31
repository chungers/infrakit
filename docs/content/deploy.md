<!--[metadata]>
+++
title = "Deploying Apps on GCP"
description = "Deploying Apps on GCP"
keywords = ["iaas, gcp"]
[menu.main]
name="Deploying Apps"
identifier="docs-apps"
weight="3"
+++
<![end-metadata]-->

# Deploying your app

## Connecting to your manager nodes

This section will walk you through connecting to your installation and deploying
applications.

## Connecting via SSH

#### Manager nodes

Once you've deployed Docker on GCP with the cli, you will see instructions on
how to connect to a manager. It will be something like:

```
OUTPUTS     VALUE
externalIp  130.211.77.165
ssh         You can ssh into the Swarm with: gcloud compute ssh --zone europe-west1-d docker-manager-1
zone        europe-west1-d
```

Follow those instructions to connect to a manager node. Any manager can be used:

    $ gcloud compute ssh --zone [zone] [manager-name]
    Welcome to Docker!

The first time you use `gcloud compute ssh` it will create an ssh key for you
and propagate it to the Swarm nodes. You can also connect to an instance [via
the cloud console] or via a [standard ssh command].

Once you are logged into the manager, you can run Docker commands on the Swarm:

    $ docker info
    $ docker node ls

You can also tunnel the Docker socket over SSH to remotely run commands on the
Swarm (requires [OpenSSH 6.7](https://lwn.net/Articles/609321/) or later):

    $ gcloud compute ssh --zone [zone] [manager-name] -- -NL localhost:2374:/var/run/docker.sock &
    $ docker -H localhost:2374 info

If you don't want to pass `-H`, you can set the `DOCKER_HOST` environment
variable to point to the localhost tunnel opening.

    $ export DOCKER_HOST=localhost:2374
    $ docker info

### Worker nodes

To be done.

## Running apps

You can now start creating containers and services:

    $ docker run hello-world

You can run websites too. Ports exposed with `-p` are automatically exposed
through the platform load balancer:

    $ docker service create --name nginx -p 80:80 nginx

Once up, connect to the site via the `externalIp` shown when you created the
stack with the cli.

### Execute docker commands in all Swarm nodes

There are cases (such as installing a volume plugin) wherein a docker command
may need to be executed in all the nodes across the cluster. You can use the
`swarm-exec` tool to achieve that.

Usage:

    $ swarm-exec {Docker command}

The following will install a test plugin in all the nodes in the cluster.

Example:

    $ swarm-exec docker plugin install --grant-all-permissions mavenugo/test-docker-netplugin

This tool internally makes use of docker global-mode service that runs a task on
each of the nodes in the cluster. This task in turn executes your docker
command. The global-mode service also guarantees that when a new node is added
to the cluster or during upgrades, a new task is executed on that node and hence
the docker command will be automatically executed.

### Distributed Application Bundles

To deploy complex multi-container apps, you can use
[distributed application bundles]. You can either run `docker deploy` to deploy
a bundle on your machine over an SSH tunnel, or copy the bundle (for example
using `scp`) to a manager node, SSH into the manager and then run
`docker deploy` (if you have multiple managers, you have to ensure that your
  session is on one that has the bundle file).

A good sample app to test application bundles is the [Docker voting app].

By default, apps deployed with bundles do not have ports publicly exposed.
Update port mappings for services, and Docker will automatically wire up the
underlying platform load balancers:

    $ docker service update --publish-add 80:80 <example-service>

### Images in private repos

To create swarm services using images in private repos, first make sure you're
authenticated and have access to the private repo, then create the service with
the `--with-registry-auth` flag (the example below assumes you're using Docker
Hub):

    $  docker login
    ...
    $  docker service create --with-registry-auth user/private-repo
    ...

This will cause swarm to cache and use the cached registry credentials when
creating containers for the service.

 [via the cloud console]: https://cloud.google.com/compute/docs/instances/connecting-to-instance#sshinbrowser
 [standard ssh command]: https://cloud.google.com/compute/docs/instances/connecting-to-instance#standardssh
 [distributed application bundles] https://github.com/docker/docker/blob/master/experimental/docker-stacks-and-bundles.md
 [Docker voting app] https://github.com/docker/example-voting-app

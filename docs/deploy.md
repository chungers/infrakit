<!--[metadata]>
+++
title = "Deploying Apps on AWS/Azure"
description = "Deploying Apps on AWS/Azure"
keywords = ["iaas, aws, azure"]
[menu.main]
name="Deploying Apps"
identifier="docs-apps"
weight="3"
+++
<![end-metadata]-->

# Deploying your app

## Connecting to your manager nodes

Once you've deployed Docker on AWS or Azure, go to the "outputs" section.
The output will show how to SSH to an SSH host. Your SSH session will be on one of the manager nodes.

    $ ssh -i <path-to-ssh-key> docker@<ssh-host>
    Welcome to Docker!

Once you are logged into the container you can run Docker commands on the swarm:

    $ docker info
    $ docker node ls

You can also tunnel the Docker socket over SSH to remotely run commands on the cluster (requires [OpenSSH 6.7](https://lwn.net/Articles/609321/) or later):

    $ ssh -NL localhost:2374:/var/run/docker.sock docker@<ssh-host> &
    $ docker -H localhost:2374 info

If you don't want to pass `-H` when using the tunnel, you can set the `DOCKER_HOST` environment variable to point to the localhost tunnel opening.

## Running apps

You can now start creating containers and services.

    $ docker run hello-world

You can run websites too. Ports exposed with `-p` are automatically exposed through the platform load balancer:

    $ docker service create --name nginx -p 80:80 nginx

Once up, find the `DefaultDNSTarget` output in either the AWS or Azure portals to access the site.

### Distributed Application Bundles

To deploy complex multi-container apps, you can use [distributed application bundles](https://github.com/docker/docker/blob/master/experimental/docker-stacks-and-bundles.md). You can either run `docker deploy` to deploy a bundle on your machine over an SSH tunnel, or copy the bundle (for example using `scp`) to a manager node, SSH into the manager and then run `docker deploy` (if you have multiple managers, you have to ensure that your session is on one that has the bundle file).

A good sample app to test application bundles is the [Docker voting app](https://github.com/docker/example-voting-app).

By default, apps deployed with bundles do not have ports publicly exposed. Update port mappings for services, and Docker will automatically wire up the underlying platform load balancers:

    docker service update --publish-add 80:80 <example-service>

### Images in private repos

To create swarm services using images in private repos, first make sure you're authenticated and have access to the private repo, then create the service with the `--with-registry-auth` flag (the example below assumes you're using Docker Hub):

    docker login
    ...
    docker service create --with-registry-auth user/private-repo
    ...

This will cause swarm to cache and use the cached registry credentials when creating containers for the service.

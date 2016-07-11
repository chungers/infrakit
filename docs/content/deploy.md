<!--[metadata]>
+++
title = "Deploying Apps on AWS/Azure"
description = "Deploying Apps on AWS/Azure"
keywords = ["iaas, aws, azure"]
[menu.iaas]
name="Deploying Apps"
identifier="docs-apps"
weight="3"
+++
<![end-metadata]-->

# Deploying your app

## Connecting to your manager nodes

Once you've created the stack, you can go to the "outputs" section in the CloudFormation stack list detail page.
The output will show how to SSH to an SSH host. Your SSH session will be on one of the manager nodes.

    $ ssh -i <path-to-ssh-key> <ssh-host-name>
    Welcome to Docker!

Once you are logged into the container you can run Docker commands on the cluster:

    $ docker swarm info
    $ docker node ls

You can also tunnel the Docker socket over SSH to remotely run commands on the cluster (requires [OpenSSH 6.7](https://lwn.net/Articles/609321/) or later):

    $ ssh -NL localhost:2374:/var/run/docker.sock docker@<ssh-host-name> &
    $ docker -H localhost:2374 info

If you don't want to pass `-H` when using the tunnel, you can set the `DOCKER_HOST` environment variable to point to the localhost tunnel opening.

## Running apps

You can now start creating containers and services.

    $ docker run hello-world

You can run websites too. Ports exposed with `-p` are automatically exposed through the platform load balancer:

    $ docker service create --name nginx -p 80:80 nginx

Once up, find the `DefaultDNSTarget` output in either the AWS or Azure portals to access the site.

To deploy complex multi-container apps, you can use [distributed application bundles](https://github.com/docker/docker/blob/master/experimental/docker-stacks.md). You can either run `docker deploy` to deploy a bundle on your machine over an SSH tunnel, or copy the bundle (for example using `scp`) to a manager node, SSH into the manager and then run `docker deploy` (if you have multiple managers, you have to ensure that your session is on one that has the bundle file).

A good sample app to test application bundles is the [Docker voting app](https://github.com/docker/example-voting-app).

By default, apps deployed with bundles do not have ports publicly exposed. Update port mappings for services, and Docker will automatically wire up the underlying platform load balancers:

    docker service update -p 80:80 <example-service>

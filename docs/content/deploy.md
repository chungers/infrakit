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

First, obtain the public IP address for a manager node (any manager node is
acceptable).

### Manager Public IP on AWS

Once you've deployed Docker on AWS, go to the "Outputs" tab for the stack in
CloudFormation.

The "Managers" output is a URL you can use to see the available manager nodes of
the cluster in your AWS console.  Once present on this page, you can see the
"Public IP" of each manager node in the table and/or "Description" tab if you
click on the instance.

![](/img/aws/managers.png)

### Manager Public IP and SSH ports on Azure

Once you've deployed Docker on Azure, go to the "Outputs" section of the resource
group deployment. The "SSH Targets" output is a URL to a blade that describes
the IP address (common across all the manager nodes) and the SSH port (unique for
each manager node) that you can use to log in to each manager node.

![](/img/azure/managers.png)

### Connecting via SSH

Obtain the public IP for the manager node and SSH in using your provided key to
begin administrating your cluster:

    $ ssh -i <path-to-ssh-key> docker@<ssh-host>
    Welcome to Docker!

In case of Azure, you also need to specify the unique port associated with a manager

    $ ssh -i <path-to-ssh-key> -p <ssh-port> docker@<ip>
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

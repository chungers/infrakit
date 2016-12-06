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

#### Connecting to worker nodes via SSH

As of Beta 13, the worker nodes also have SSH enabled. By default SSH access is not allowed to the worker nodes from the public internet. To access the worker nodes, you will need to first connect to a manager node (see above). On the manager node you can then ssh to the worker node, over the private network. Make sure you have `ssh agent forwarding` enabled (see below). If you run the `docker node ls` command you can see the full list of nodes in your swarm. You can then `ssh docker@<worker-host>` to have access to that node.

```
$ docker node ls
ID                           HOSTNAME                                     STATUS  AVAILABILITY  MANAGER STATUS
a3d4vdn9b277p7bszd0lz8grp *  ip-172-31-31-40.us-east-2.compute.internal   Ready   Active        Reachable
axv7lb0quc6lhwl41iixbnbev    ip-172-31-22-131.us-east-2.compute.internal  Ready   Active
kpjqylqkm3c5sr2lfwn0ov063    ip-172-31-8-210.us-east-2.compute.internal   Ready   Active        Reachable
rb52lqiku01b1o10grx6qiu9x    ip-172-31-8-211.us-east-2.compute.internal   Ready   Active        Leader
xxnoctzqji2k3l5rz7zkvdevp    ip-172-31-13-108.us-east-2.compute.internal  Ready   Active
```

Connecting to a worker node from the list above.

```
$ ssh docker@ip-172-31-22-131.us-east-2.compute.internal
```

##### SSH agent forwarding

SSH agent forwarding allows you to forward along your ssh keys when connecting from one node to another. This eliminates the need for installing your private key on all nodes you might want to connect from.

If your haven't added your ssh key to the `ssh-agent` you will also need to do this first.

See the keys in the agent already.

```
$ ssh-add -L
```

If you don't see your key, add it like this.

```
$ ssh-add ~/.ssh/your_key
```

On Mac OS X, the `ssh-agent` will forget this key, once it gets restarted. But you can import your SSH key into your Keychain like this. This will have your key survive restarts.

```
$ ssh-add -K ~/.ssh/your_key
```

If you don't have agent forwarding turned on by default, you won't be able to SSH into the worker, and you will need to enable it when connecting to the manager. You can enable using the `-A` flag for the ssh command.

Connecting to the Manager.
```
$ ssh -A docker@<manager ip>
```

To always have it turned on, you can edit your ssh config file (`/etc/ssh_config`, `~/.ssh/config`, etc) and make sure it says `ForwardAgent yes`

For example, on Mac OS X the file is `/etc/ssh_config` and you need the following.

```
Host *
  ForwardAgent yes
```


## Running apps

You can now start creating containers and services.

    $ docker run hello-world

You can run websites too. Ports exposed with `-p` are automatically exposed through the platform load balancer:

    $ docker service create --name nginx -p 80:80 nginx

Once up, find the `DefaultDNSTarget` output in either the AWS or Azure portals to access the site.

### Execute docker commands in all swarm nodes

There are cases (such as installing a volume plugin) wherein a docker command may need to be executed in all the nodes across the cluster. You can use the `swarm-exec` tool to achieve that.

Usage : `swarm-exec {Docker command}`

The following will install a test plugin in all the nodes in the cluster

Example : `swarm-exec docker plugin install --grant-all-permissions mavenugo/test-docker-netplugin`

This tool internally makes use of docker global-mode service that runs a task on each of the nodes in the cluster. This task in turn executes your docker command. The global-mode service also guarantees that when a new node is added to the cluster or during upgrades, a new task is executed on that node and hence the docker command will be automatically executed.

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

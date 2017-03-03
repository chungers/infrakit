<!--[metadata]>
+++
title = "Docker for GCP Setup"
description = "Docker for GCP Setup"
keywords = ["iaas, gcp"]
[menu.main]
identifier="docs-gcp-index"
parent = "docs-gcp"
name = "Setup & Prerequisites"
weight="100"
+++
<![end-metadata]-->

# Docker for GCP Setup

## Prerequisites

- Access to a Google Cloud project with those Api enabled:
  - [Google Cloud Deployment Manager V2 API]
  - [Google Cloud RuntimeConfig API]
- Make sure that you have enough capacity for the Swarm that you want to build,
and won't go over any of your limits.
- Optional: install [gcloud] SDK. It's not mandatory but makes
interacting with your project easier.

Once you have all of the above you are ready to move onto the next step.

## Configuration

Docker for GCP is installed with a [Deployment Manager] template that configures
Docker in Swarm-mode, running on instances backed by custom images. You can use
the [gcloud] SDK either from your machine or from [Google Cloud Shell].
The later is easier since it doesn't require installing any tool on your
machine.

### Installing from Google Cloud Shell

Open your browser, connect to the [GCP Console], start [Google Cloud Shell] and
type this command with property values suited to your needs. For example:

    $ gcloud deployment-manager deployments create docker \
        --config https://docker-for-gcp-templates.storage.googleapis.com/v8/Docker.jinja \
        --properties managerCount:3,workerCount:1

### Installing with the CLI

If you prefer to not use [Google Cloud Shell], you will need to install
[gcloud], then run those commands:

    $ gcloud init --skip-diagnostics
    $ gcloud deployment-manager deployments create docker \
        --config https://docker-for-gcp-templates.storage.googleapis.com/v8/Docker.jinja \
        --properties managerCount:3,workerCount:1

### Stack name

The name `docker` can be replaced with another name that will uniquely identify
your stack in a GCP project. The stack will have its own instances, network,
external ip address and load balancer, so that multiple stacks can co-exist
within a single project.

### Configuration options

This above example shows how to configure the number of Swarm-mode managers and
workers. There are more options that you can provide to customized the Swarm.

#### managerCount

The number of [Managers] in your Swarm. You can pick either 1, 3 or 5 managers.
We only recommend 1 manager for testing and dev setups. There are no failover
guarantees with 1 manager — if the manager fails the Swarm will go down as well.
Additionally, upgrading single-manager Swarms is not currently guaranteed to
succeed.

We recommend at least 3 managers, and if you have a lot of workers, you should
pick 5 managers.

When choosing manager count, consider the level of durability you need:

| # of managers  | # of tolerated failures |
| -------------- | ----------------------- |
|             1  |                      0  |
|             3  |                      1  |
|             5  |                      2  |

#### workerCount

The number of [Workers] in your Swarm

#### zone

The [Zone] to which the nodes are attached. The default value is `us-central1-f`.

#### managerMachineType

The [machine type] for your Manager nodes. The default value is `g1-small`.
If you're testing, `g1-small` instances are fine.

#### workerMachineType

The [machine type] for your Worker nodes. The default value is `g1-small`.
If you're testing, `g1-small` instances are fine.

#### managerDiskSize

The size of the Manager boot disks in Mb. The default value is `100`.

#### workerDiskSize

The size of the Worker boot disks in Mb. The default value is `100`.

#### managerDiskType

The [Disk Type] used by Managers. The default value is `pd-standard`.

#### workerDiskType

The [Disk Type] used by Workers. The default value is `pd-standard`.

#### enableSystemPrune

Enable if you want Docker for GCP to automatically cleanup unused space on your
Swarm nodes.

When enabled, `docker system prune` will run staggered every day, starting at
1:42AM UTC on both workers and managers. The prune times are staggered slightly
so that not all nodes will be pruned at the same time. This limits resource
spikes on the Swarm.

Pruning removes the following:

 - All stopped containers
 - All volumes not used by at least one container
 - All dangling images
 - All unused networks

## How it works

Docker for GCP starts with a [Deployment Manager] template that will create
everything that you need from scratch. There are only a few prerequisites that
are listed above.

It first starts off by creating a new network along with subnets and firewall
rules. Once the networking is set up, it will create one machine that will be
the swarm's first manager. This manager starts [Infrakit] that will take care of
spawning Manager and Worker nodes. At this point you will have x number of
managers and y number of workers in your swarm, that are ready to handle your
application deployments. See the [deployment] docs for your next steps.

A load balancer is also set up to help with routing traffic to your Swarm.

## System containers

Each node will have a few system containers running on them to help run your
Swarm. In order for everything to run smoothly, please keep those containers
running, and don't make any changes. If you make any changes, Docker for GCP
will not work correctly.

 [Google Cloud Deployment Manager V2 API]: https://console.developers.google.com/apis/api/deploymentmanager-json.googleapis.com/overview
 [Google Cloud RuntimeConfig API]: https://console.developers.google.com/apis/api/runtimeconfig.googleapis.com/overview
 [gcloud]: https://cloud.google.com/sdk/downloads
 [Deployment Manager]: https://cloud.google.com/deployment-manager/docs/
 [GCP Console]: https://console.cloud.google.com/home/dashboard
 [Google Cloud Shell]: https://cloud.google.com/shell/docs/quickstart#start_cloud_shell
 [Managers]: https://docs.docker.com/engine/swarm/key-concepts/#/what-is-a-node
 [Workers]: https://docs.docker.com/engine/swarm/key-concepts/#/what-is-a-node
 [Zone]: https://cloud.google.com/compute/docs/regions-zones/viewing-regions-zones
 [machine type]: https://cloud.google.com/compute/docs/machine-types
 [Disk Type]: https://cloud.google.com/compute/docs/disks/#pdspecs
 [Infrakit]: https://github.com/docker/infrakit
 [deployment]: ../deploy.md

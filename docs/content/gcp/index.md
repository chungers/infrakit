<!--[metadata]>
+++
title = "Docker for GCP"
description = "Docker for GCP"
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
  - [Google Cloud Deployment Manager V2 API](https://console.developers.google.com/apis/api/deploymentmanager-json.googleapis.com/overview?project=docker4x&duration=PT1H)
  - [Google Cloud RuntimeConfig API](https://console.developers.google.com/apis/api/runtimeconfig.googleapis.com/overview?project=docker4x)
- Make sure that you have enough capacity for the swarm that you want to build, and won't go over any of your limits.

Once you have all of the above you are ready to move onto the next step.

## Configuration

Docker for GCP is installed with a Deployment Manager template that configures Docker in swarm-mode, running on instances backed custom images. There are two ways you can deploy Docker for GCP. You can use the `gcloud` CLI from any machine or from Google Cloud Shell.

### Installing with the CLI

Here is an example of how to use the CLI:

```
$ gcloud init --skip-diagnostics
$ gcloud deployment-manager deployments create docker-stack \
    --config https://storage.googleapis.com/docker-for-gcp-templates/gcp-v1.13.0-rc6-beta16/swarm.jinja \
    --properties managerCount:3,workerCount:2,managerMachineType:g1-small,workerMachineType:g1-small
```

If you run the second command from Google Cloud Shell, you don't need the `init` command since
you are already authenticated to connect to Cloud Shell.

### Configuration options

This above example shows how to configure the number of Managers and Workers, as well as the type of machines to use.
There are more options that you can provide to customized the swarm.

#### managerMachineType
The [machine type](https://cloud.google.com/compute/docs/machine-types) for your Manager nodes.

#### workerMachineType
The [machine type](https://cloud.google.com/compute/docs/machine-types) for your Worker nodes.

#### managerCount
The number of Managers in your swarm. You can pick either 1, 3 or 5 managers. We only recommend 1 manager for testing and dev setups. There are no failover guarantees with 1 manager — if the single manager fails the swarm will go down as well. Additionally, upgrading single-manager swarms is not currently guaranteed to succeed.

We recommend at least 3 managers, and if you have a lot of workers, you should pick 5 managers.

#### workerCount
The number of Workers in your swarm

### zone
The [zone](https://cloud.google.com/compute/docs/regions-zones/viewing-regions-zones) to which the nodes are attached.

#### enableSystemPrune

Enable if you want Docker for GCP to automatically cleanup unused space on your swarm nodes.

When enabled, `docker system prune` will run staggered every day, starting at 1:42AM UTC on both workers and managers. The prune times are staggered slightly so that not all nodes will be pruned at the same time. This limits resource spikes on the swarm.

Pruning removes the following:
- All stopped containers
- All volumes not used by at least one container
- All dangling images
- All unused networks

## How it works

Docker for GCP starts with a Deployment Manager template that will create everything that you need from scratch. There are only a few prerequisites that are listed above.

It first starts off by creating a new network along with subnets and firewall rules. Once the networking is set up, it will create one machine that will be the swarm's first manager. This manager starts [Infrakit](https://github.com/docker/infrakit) that will take care of spawning Managers and Workers nodes. At this point you will have x number of managers and y number of workers in your swarm, that are ready to handle your application deployments. See the [deployment](../deploy.md) docs for your next steps.

A load balancer is set up to help with routing traffic to your swarm.

## System containers

Each node will have a few system containers running on them to help run your swarm cluster. In order for everything to run smoothly, please keep those containers running, and don't make any changes. If you make any changes, Docker for GCP will not work correctly.

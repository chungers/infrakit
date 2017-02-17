<!--[metadata]>
+++
title = "Docker for GCP Upgrade"
description = "Docker for GCP Upgrade"
keywords = ["iaas, gcp"]
[menu.main]
identifier="docs-gcp-upgrade"
parent = "docs-gcp"
name = "Upgrading"
weight="300"
+++
<![end-metadata]-->

# Docker for GCP Upgrades

_Warning_: Upgrades are currently not supported. 
Performing an upgrade on an existing cluster may leave some of the nodes in a broken state. 

Docker for GCP will support upgrading from one beta version to the next. Upgrades
are done by applying a new version of the Deployment Manager template that
powers Docker for GCP. Depending on changes in the next version, an upgrade
involves:

 - Changing the Base Disk Image backing manager and worker nodes
   (the Docker engine ships in the image)
 - Upgrading service containers
 - Changing the resource setup

To be notified of updates, submit your email address at
[https://beta.docker.com/].

## Prerequisites

 - We recommend only attempting upgrades of swarms with at least 3 managers.
   A 1-manager swarm may not be able to maintain quorum during the upgrade
 - Upgrades are only supported from one version to the next version, for example
   v11 to v12. Skipping a version during an upgrade is not supported. For
   example, upgrading from v10 to v12 is not supported.
 - Downgrades are not supported.

## Upgrading

If you submit your email address at [https://beta.docker.com/] Docker will
notify you of new releases by email. New releases are also posted on the
[Release Notes] page.

To initiate an update, use [gcloud] cli to initiate a stack update. Use the
template URL for the new release. This will initiate a rolling upgrade of the
Docker Swarm, and service state will be maintained during and after the upgrade.
Appropriately scaled services should not experience downtime during an upgrade.

    $ gcloud deployment-manager deployments update docker \
        --config https://docker-for-gcp-templates.storage.googleapis.com/v[NEW-VERSION]/Docker.jinja \
        --properties managerCount:3,workerCount:5

_Warning_: If you created your deployment with non default settings (node count,
machine type...), you must reuse the same settings in the `update` command that
you used with the `create` command along with the updated template url.

Note that single containers started (for example) with `docker run -d` are
**not** preserved during an upgrade. This is because they're not Docker Swarm
objects, but are known only to the individual Docker engines.

## Changing instance sizes and other template parameters

In addition to upgrading Docker for GCP from one version to the next you can
also change template parameters such as worker count and instance type.
Changing manager count is **not** supported.

 [https://beta.docker.com/]: https://beta.docker.com/
 [Release Notes]: https://beta.docker.com/docs/aws/release-notes/
 [gcloud]: https://cloud.google.com/sdk/downloads

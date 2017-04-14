<!--[metadata]>
+++
title = "Docker for GCP Release notes"
description = "Docker for GCP Release notes"
keywords = ["iaas, gcp"]
[menu.main]
identifier="gcp-release-notes"
parent = "docs-gcp"
name = "Release Notes"
weight="400"
+++
<![end-metadata]-->

# Release notes

## v9

+ Managers now have two disks. One for the kernel, one for the data. This makes
the upgrades much smoother.
+ Fixed the channel name. Should be `edge`, not `beta`.
+ System docker images are now read from disk rather than pulled from the network.
+ Docker 17.05.0-ce-rc1

## v8

+ Fix the image pruning. It was not possible to use `enableSystemPrune` and the
cron that actually calls `docker prune` was broken.
+ Add documentation to delete a stack.
+ All the gcloud commands in the docs and the outputs should reference the
project. eg, `gcloud compute ssh --project [project] --zone [zone] [manager-name]`.
+ Access to ssh can be restricted to a class of source addresses.
+ Improve stack scaling/upgrade.
+ Base disk images are not anymore public.
+ Docker 17.03.0-ce

## v7

+ First version opened to chosen Beta testers
+ Docker 17.03.0-ce-rc1
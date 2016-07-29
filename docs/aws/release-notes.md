<!--[metadata]>
+++
aliases = [
"/aws-release-notes/"
]
title = "Docker for AWS Release notes"
description = "Docker for AWS Release notes"
keywords = ["iaas, aws"]
[menu.main]
identifier="aws-release-notes"
parent = "docs-aws"
name = "Release Notes"
weight="400"
+++
<![end-metadata]-->

# Docker for AWS Release notes

## 1.12.0-beta4

Release date: 7/28/2016

### New

 * Docker Engine upgraded to 1.12.0

### Errata

 * Upgrading from previous Docker for AWS versions to 1.12.0-beta4 is not possible because of RC-incompatibilities between Docker 1.12.0 release candidate 5 and previous release candidates.

## 1.12.0-rc5-beta3

(internal release)

## 1.12.0-rc4-beta2

Release date: 7/13/2016

### New

 * Docker Engine upgraded to 1.12.0-rc4
 * EC2 instance tags
 * Beta Docker for AWS sends anonymous analytics

### Errata
 * When upgrading, old Docker nodes may not be removed from the swarm and show up when running `docker node ls`. Marooned nodes can be removed with `docker node rm`

## 1.12.0-rc3-beta1

### New

 * First release of Docker for AWS!
 * CloudFormation based installer
 * ELB integration for running public-facing services
 * Swarm access with SSH
 * Worker scaling using AWS ASG

### Errata

 * To assist with debugging, the Docker Engine API is available internally in the AWS VPC on TCP port 2375. These ports cannot be accessed from outside the cluster, but could be used from within the cluster to obtain privileged access on other cluster nodes. In future releases, direct remote access to the Docker API will not be available.
 * Likewise, swarm-mode is configured to auto-accept both manager and worker nodes inside the VPC. This policy will be changed to be more restrictive by default in the future.

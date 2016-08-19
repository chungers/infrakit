<!--[metadata]>
+++
title = "Docker for Azure Release notes"
description = "Docker for Azure Release notes"
keywords = ["iaas, Azure"]
[menu.main]
identifier="azure-release-notes"
parent = "docs-azure"
name = "Release Notes"
weight="400"
+++
<![end-metadata]-->

# Docker for Azure Release notes

## 1.12.1-beta5

Release date: 8/19/2016

### New

 * Docker Engine upgraded to 1.12.1

### Errata

 * To assist with debugging, the Docker Engine API is available internally in the Azure VPC on TCP port 2375. These ports cannot be accessed from outside the cluster, but could be used from within the cluster to obtain privileged access on other cluster nodes. In future releases, direct remote access to the Docker API will not be available.

## 1.12.0-beta4

Release date: 8/9/2016

### New

 * First release

### Errata

 * To assist with debugging, the Docker Engine API is available internally in the Azure VPC on TCP port 2375. These ports cannot be accessed from outside the cluster, but could be used from within the cluster to obtain privileged access on other cluster nodes. In future releases, direct remote access to the Docker API will not be available.

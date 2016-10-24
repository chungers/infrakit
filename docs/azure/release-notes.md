<!--[metadata]>
+++
aliases = [
"/azure-release-notes/"
]
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

## 1.12.2-beta9

Release date: 10/17/2016

<a href="https://portal.azure.com/#create/Microsoft.Template/uri/https%3A%2F%2Fdocker-for-azure.s3.amazonaws.com%2Fazure%2Fbeta%2Fazure-v1.12.2-beta9.json" target="_blank" id="azure-deploy">![Docker for Azure](https://gallery.mailchimp.com/761fa9756d4209ea04a811254/images/f9aab976-fd63-4e64-bb66-5e57e1ffd9c1.png)</a>

### New

- Docker Engine upgraded to Docker 1.12.2
- Manager behind its own LB
- Added sudo support to the shell container on manager nodes

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

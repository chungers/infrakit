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

## 1.12.0-rc4-beta2

Release date: 7/13/2016

### New

 * Docker Engine upgraded to 1.12.0


### Errata

 * When upgrading, old Docker nodes may not be removed from the swarm and show up when running `docker node ls`. Marooned nodes can be removed with `docker node rm`

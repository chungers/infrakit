<!--[metadata]>
+++
title = "Docker for AWS Release notes"
description = "Docker for AWS Release notes"
keywords = ["iaas, aws"]
[menu.iaas]
identifier="aws-release-notes"
weight="100"
+++
<![end-metadata]-->

# Docker for AWS Release notes

## 1.12.0-rc3-beta1

### New

 * First release of Docker for AWS!
 * Cloudformation-based installer
 * ELB integration for running public-facing services
 * Swarm access with SSH
 * Worker scaling using AWS ASG

### Errata

 * The Docker Engine API is available internally in the AWS VPC on TCP port 2375. This is not immediately exploitable, but if a container running on the swarm is compromised the weakness can be used to get additional privileges. In future releases, workers and manager Docker engine APIs will not be bound to network interfaces.
 * Swarm-mode is configured to auto-accept both manager and worker nodes. This is not immediately exploitable, but if a container running on the swarm is compromised an attacker can masquerade as a joining node and get additional privileges. In future releases the swarm accept policy will be changed to not auto-accept workers or managers.

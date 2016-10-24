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

## 1.12.2-beta9

Release date: 10/12/2016

<a href="https://console.aws.amazon.com/cloudformation/home#/stacks/new?stackName=Docker&templateURL=https://docker-for-aws.s3.amazonaws.com/aws/beta/aws-v1.12.2-beta9.json" data-rel="Beta-9" target="_blank" id="aws-deploy">![Docker for AWS](https://gallery.mailchimp.com/761fa9756d4209ea04a811254/images/da458f6b-3c2c-414b-9f3e-e5819ad3761b.png)</a>

### New

- Docker Engine upgraded to Docker 1.12.2
- Can better handle scaling swarm nodes down and back up again
- Container logs are now sent to CloudWatch
- Added a diagnostic command (docker-diagnose), to more easily send us diagnostic information incase of errors for troubleshooting
- Added sudo support to the shell container on manager nodes
- Change SQS default message timeout to 12 hours from 4 days
- Added support for region 'ap-south-1': Asia Pacific (Mumbai)

### Deprecated:
- Port 2375 will be closed in next release. If you relay on this being open, please plan accordingly.

## 1.12.2-RC3-beta8

Release date: 10/06/2016

 * Docker Engine upgraded to 1.12.2-RC3

## 1.12.2-RC2-beta7

Release date: 10/04/2016

 * Docker Engine upgraded to 1.12.2-RC2

## 1.12.2-RC1-beta6

Release date: 9/29/2016

### New

 * Docker Engine upgraded to 1.12.2-RC1


## 1.12.1-beta5

Release date: 8/18/2016

### New

 * Docker Engine upgraded to 1.12.1

### Errata

 * Upgrading from previous Docker for AWS versions to 1.12.0-beta4 is not possible because of RC-incompatibilities between Docker 1.12.0 release candidate 5 and previous release candidates.

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

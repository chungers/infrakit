<!--[metadata]>
+++
title = "Docker for AWS"
description = "Docker for AWS"
keywords = ["iaas, aws, azure"]
[menu.main]
identifier="docs-aws-upgrade"
parent = "docs-aws"
name = "Upgrading"
weight="300"
+++
<![end-metadata]-->

# Docker for AWS Upgrades
There is currently limited support for upgrades, we will be improving this in future releases. The way upgrades work is as follows. We will release a new CloudFormation template, and you will update your CloudFormation stack with the new CloudFormation template. CloudFormation will look at your current stack, and see what is different, and let you know what it is about to do. If you confirm the changes it will start to update your stack.

If there is a change to the AMI, then one by one the old nodes will be shut down, and replaced by newer nodes. How long this takes will depend on how many nodes you have. We do a slow rolling update, so the more nodes, the longer it will take. Eventually when complete, all of the older nodes will be gone, and replaced with a completely new set of nodes.

Since we are doing a slow rolling upgrade, the services that are running on a node that is getting shut down, will be rescheduled by swarm, and put on to a new healthy node. If your service is properly scaled it should not notice any downtime.

Since this feature is still very new, there is a chance things could go wrong, so we don't recommend using Docker for AWS for production critical workloads at this time.


## Upgrading Docker and changing instance sizes
In the AWS Console, find your CloudFormation stack and select "Update stack". Use the CloudFormation template link. This will let you change the input parameters for the template. AWS will summarize the proposed changes, whether that's changing the AMIs to upgrade Docker or to change instance sizes.

Docker will ensure that upgrade and instance size changes are handled with minimal impact to running apps.

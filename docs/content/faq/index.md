<!--[metadata]>
+++
title = "Docker for AWS & Azure FAQ"
description = "Docker for AWS & Azure FAQ"
keywords = ["iaas, aws, azure, faq"]
[menu.main]
identifier="faq-index"
parent = "docs-aws-azure-faq"
name = "Overview"
weight="110"
+++
<![end-metadata]-->

# FAQ

## How stable are Docker for AWS and Docker for Azure

Docker for AWS and Azure are stable enough for development and testing, but we currently don't recommend use with production workloads.

## I have a problem/bug where do I report it?

Send an email to <docker-for-iaas@docker.com> or post to the [Docker for AWS](https://forums.docker.com/c/docker-for-aws) or the [Docker for Azure](https://forums.docker.com/c/docker-for-azure) forums.

In AWS (coming to Azure soon), if your stack/resource group is misbehaving, please run the following diagnostic tool from one of the managers - this will collect your docker logs and send them to Docker:

```
$ docker-diagnose
OK hostname=manager1
OK hostname=worker1
OK hostname=worker2
Done requesting diagnostics.
Your diagnostics session ID is 1234567890-xxxxxxxxxxxxxx
Please provide this session ID to the maintainer debugging your issue.
```

_Please note that your output will be slightly different from the above, depending on your swarm configuration_

## Analytics

The beta versions of Docker for AWS and Azure send anonymized analytics to Docker. These analytics are used to monitor beta adoption and are critical to improve Docker for AWS and Azure.

## How to run administrative commands?

By default when you SSH into a manager, you will be logged in as the regular username: `docker` - It is possible however to run commands with elevated privileges by using `sudo`.
For example to ping one of the nodes, after finding its IP via the Azure/AWS portal (e.g. 10.0.0.4), you could run:
```
$ sudo ping 10.0.0.4
``` 

Note that access to Docker for AWS and Azure happens through a shell container that itself runs on Docker.
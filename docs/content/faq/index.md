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

## How long will it take before I get accepted into the private beta?
Docker for AWS is built on top of Docker 1.12, but as with all Beta, things are still changing, which means things can break between release candidates.

We are currently rolling it out slowly to make sure everything is working as it should. This is to ensure that if there are any issues we limit the number of people that are affected.

## How stable is Docker for AWS and Docker for Azure
We feel it is fairly stable for development and testing, but since things are consistently changing, we currently don't recommend using it for production workloads at this time.

## I have a problem/bug where do I report it?
Send an email to <docker-for-iaas@docker.com> or use the [Docker for AWS Forum](https://forums.docker.com/c/docker-for-aws) or the [Docker for Azure Forum](https://forums.docker.com/c/docker-for-azure)

In AWS (coming to Azure soon), if your stack/resource group is misbehaving, please run the following diagnostic tool from one of the managers; this will collect your docker logs and send them to us:

```
$ docker-diagnose
OK hostname=manager1
OK hostname=worker1
OK hostname=worker2
Done requesting diagnostics.
Your diagnostics session ID is 1234567890-xxxxxxxxxxxxxx
Please provide this session ID to the maintainer debugging your issue.
```

_Please note that your output will be slightly different from the above and will reflect your nodes_


## Analytics
The beta versions of Docker for AWS and Azure send anonymized analytics to Docker. These analytics are used to monitor beta adoption and are critical to improve Docker for AWS and Azure.

## How to run administrative commands?
By default when you SSH into the manager, you will be logged in as the regular username: `docker` - It is possible however to run commands with elevated privileges by using `sudo`.
For example to ping one of the nodes, after finding its IP via the Azure/AWS portal (e.g. 10.0.0.4), you could run:
```
$ sudo ping 10.0.0.4
``` 

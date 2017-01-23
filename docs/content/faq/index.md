<!--[metadata]>
+++
title = "Docker for GCP FAQ"
description = "Docker for GCP FAQ"
keywords = ["iaas, GCP, faq"]
[menu.main]
identifier="faq-index"
parent = "docs-gcp-faq"
name = "Overview"
weight="110"
+++
<![end-metadata]-->

# FAQ

## How stable are Docker for GCP

Docker for GCP are stable enough for development and testing, but we currently don't recommend use with production workloads.

## I have a problem/bug where do I report it?

Send an email to <docker-for-iaas@docker.com> or post to the [Docker for GCP](https://forums.docker.com/c/docker-for-gcp) forums.

In GCP, if your stack/resource group is misbehaving, please run the following diagnostic tool from one of the managers - this will collect your docker logs and send them to Docker:

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

The beta versions of Docker for GCP send anonymized analytics to Docker. These analytics are used to monitor beta adoption and are critical to improve Docker for GCP.

## How to run administrative commands?

By default when you SSH into a manager, you will be logged in as the regular username: `docker` - It is possible however to run commands with elevated privileges by using `sudo`.
For example to ping one of the nodes, after finding its IP via the GCP portal (e.g. 10.0.0.4), you could run:
```
$ sudo ping 10.0.0.4
``` 

Note that access to Docker for GCP happens through a shell container that itself runs on Docker.
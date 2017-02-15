<!--[metadata]>
+++
title = "Docker for GCP FAQ"
description = "Docker for GCP FAQ"
keywords = ["iaas, gcp, faq"]
[menu.main]
identifier="faq-index"
parent = "docs-gcp-faq"
name = "Overview"
weight="110"
+++
<![end-metadata]-->

# FAQ

## How stable is Docker for GCP

Docker for GCP is stable enough for development and testing, but we currently
don't recommend use with production workloads.

## Can I use my own Disk Image?

No, at this time we only support the default Docker Disk Image.

## Where do I report problems or bugs?

Send an email to <docker-for-iaas@docker.com> or post to the [Docker for GCP]
forums.

If your stack is misbehaving, please run the following diagnostic tool from one
of the managers - this will collect your docker logs and send them to Docker:

```bash
$ docker-diagnose
OK hostname=manager1
OK hostname=worker1
OK hostname=worker2
Done requesting diagnostics.
Your diagnostics session ID is 1234567890-xxxxxxxxxxxxxx
Please provide this session ID to the maintainer debugging your issue.
```

> **Note**: Your output will be slightly different from the above, depending on your swarm configuration.

## Metrics

Docker for GCP sends anonymized minimal metrics to Docker (heartbeat). These
metrics are used to monitor adoption and are critical to improve Docker for GCP.


 [Docker for GCP] https://forums.docker.com/c/docker-for-gcp

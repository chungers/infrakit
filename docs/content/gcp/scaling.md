<!--[metadata]>
+++
title = "Docker for GCP"
description = "Docker for GCP"
keywords = ["iaas, gcp"]
[menu.main]
identifier="docs-gcp-scaling"
parent = "docs-gcp"
name = "Scaling"
weight="200"
+++
<![end-metadata]-->

# Modifying Docker install on GCP

## Scaling workers

You can scale the worker count using the deployment manager. Docker will automatically join or remove new instances to the Swarm.
To achieve that, you need to "update" your stack, and change the number of workers in the Deployment Manager template.

Changing manager count live is _not_ currently supported.

Here is an example of how to use the CLI:

```
$ gcloud deployment-manager deployments create docker-stack \
    --config https://storage.googleapis.com/docker-for-gcp-templates/gcp-v1.13.0-rc6-beta16/swarm.jinja \
    --properties managerCount:3,workerCount:5,managerMachineType:g1-small,workerMachineType:g1-small
```

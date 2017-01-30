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

Worker machine type can also be changed. But changing manager count live is _not_ currently supported.

Here is an example of how to use the CLI:

```
$ gcloud deployment-manager deployments update docker-stack \
    --config https://docker-for-gcp-templates.storage.googleapis.com/v2/Docker.jinja \
    --properties managerCount:3,workerCount:5,managerMachineType:g1-small,workerMachineType:g1-small
```

_Warning_: If you created your deployment with non default settings, you must
reuse the same settings in the `update` command that you used with the `create`
command. Only the `workerCount` or `workerMachineType` properties should be changed.

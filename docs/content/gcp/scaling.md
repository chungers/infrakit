<!--[metadata]>
+++
title = "Docker for GCP Scaling"
description = "Docker for GCP Scaling"
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

You can scale the worker count using the [gcloud] CLI. Docker will
automatically join or remove new instances to the Swarm. To achieve that, you
need to "update" your stack, and change the number of workers passed to the
Deployment Manager template.

Worker machine type can also be changed. But changing manager count live is
_not_ currently supported.

Here is an example that sets the number of workers to `5`:

    $ gcloud deployment-manager deployments update docker \
        --config https://docker-for-gcp-templates.storage.googleapis.com/v2/Docker.jinja \
        --properties managerCount:3,workerCount:5

_Warning_: If you created your deployment with non default settings (node count,
machine type...), you must reuse the same settings in the `update` command that
you used with the `create` command along with the properties you want to
actually change.

 [gcloud]: https://cloud.google.com/sdk/downloads

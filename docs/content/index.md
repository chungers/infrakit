<!--[metadata]>
+++
title = "Docker for GCP"
description = "Docker for GCP"
keywords = ["iaas, gcp"]
[menu.main]
identifier="docs"
name = "Getting Started"
weight="1"
+++
<![end-metadata]-->

#### Looking for Docker for AWS or Azure? <a href="https://docs.docker.com/engine/installation/" target="_blank">Click here</a>

# Docker for GCP beta

Docker for GCP let you quickly set up and configure a working Docker 1.13 swarm-mode install on Google Cloud Platform.

Docker for GCP is private beta ([sign up](https://beta.docker.com/gcp) for access). It is free to use (GCP will charge for resource use).

To deploy Docker for GCP, find the latest release in the [Release Notes](gcp/release-notes.md).

## What to know before installing

When setting up Docker for GCP, you'll have to select manager and worker counts and instance sizes. If you're testing, 1 manager and `g1-small` instances are fine. Both worker count and worker and manager instance size can be changed later, but manager count should not be modified after setup.

When choosing manager count, consider the level of durability you need:

| # of managers  | # of tolerated failures |
| ------------- | ------------- |
| 1  | 0  |
| 3  | 1  |
| 5  | 2  |

For more details, check out the rest of the documentation:

 * [Docker for GCP](gcp/index.md)
 * [Deploying your Apps](deploy.md)
 * [Docker for GCP release notes](gcp/release-notes.md)

<p style="margin-bottom:50px">&nbsp;</p>

## Getting help

Reach out to <docker-for-iaas@docker.com> with questions, comments and feedback. The forums are available for public discussions:

* For GCP Help, use the [Docker for GCP Forum](https://forums.docker.com/c/docker-for-gcp)

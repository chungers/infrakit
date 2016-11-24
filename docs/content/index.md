<!--[metadata]>
+++
title = "Docker for AWS and Azure"
description = "Docker for AWS and Azure"
keywords = ["iaas, aws, azure"]
[menu.main]
identifier="docs"
name = "Getting Started"
weight="1"
+++
<![end-metadata]-->

#### Looking for Docker for Mac or Windows? <a href="https://docs.docker.com/" target="_blank">Click here</a>

# Docker for AWS and Docker for Azure beta

Docker for AWS and Azure let you quickly set up and configure a working Docker 1.13 swarm-mode install on Amazon Web Services and on Azure.

Docker for AWS is public beta while Docker for Azure is still in private beta ([sign up](https://beta.docker.com/azure) for access). Both are free to use (AWS and Azure will charge for resource use).

To deploy Docker for AWS, find the latest release in the [Release Notes](aws/release-notes.md).

## What to know before installing

When setting up Docker for AWS or Azure, you'll be prompted to select manager and worker counts and instance sizes. If you're testing, 1 manager and `small` instances are fine. Both worker count and worker and manager instance size can be changed later, but manager count should not be modified after setup.

When choosing manager count, consider the level of durability you need:

| # of managers  | # of tolerated failures |
| ------------- | ------------- |
| 1  | 0  |
| 3  | 1  |
| 5  | 2  |

For more details, check out the rest of the documentation:

 * [Docker for AWS](aws/index.md)
 * [Docker for Azure](azure/index.md)
 * [Deploying your Apps](deploy.md)
 * [Docker for AWS release notes](aws/release-notes.md)
 * [Docker for Azure release notes](azure/release-notes.md)

<p style="margin-bottom:50px">&nbsp;</p>

## Getting help

Reach out to <docker-for-iaas@docker.com> with questions, comments and feedback. The forums are available for public discussions:

* For AWS Help, use the [Docker for AWS Forum](https://forums.docker.com/c/docker-for-aws)
* For Azure Help, use the [Docker for Azure Forum](https://forums.docker.com/c/docker-for-azure)

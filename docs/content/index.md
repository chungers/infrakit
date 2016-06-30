<!--[metadata]>
+++
title = "Docker for AWS and Azure"
description = "Docker for AWS and Azure"
keywords = ["iaas, aws, azure"]
[menu.iaas]
identifier="docs"
name = "Getting Started"
weight="1"
+++
<![end-metadata]-->

#### Looking for Docker for Mac or Windows? <a href="https://docs.docker.com/" target="_blank">Click here</a>

# Docker for AWS and Docker for Azure beta

Docker for AWS and Azure are the best ways to install, configure and maintain Docker deployments on AWS and Azure. Docker for AWS and Azure both configure Docker 1.12  with swarm-mode enabled out of the box.

Docker for AWS and Azure are available in private beta for testing. They’re free to use (AWS and Azure will charge for resource use).

Sign up for the beta on [beta.docker.com](https://beta.docker.com/).

## What to know before installing

When setting up Docker for AWS or Azure, you'll be prompted to select manager and worker counts and instance sizes. If you're testing, 1 manager and `t2.micro` instances is fine. Both worker count and worker and manager instance size can be changed later, but manager count should not be modified.

When choosing manager count, consider the level of durability you need:

| # of managers  | # of tolerated failures |
| ------------- | ------------- |
| 1  | 0  |
| 3  | 1  |
| 5  | 2  |

For more details, check out the rest of the documentation:

 * [Docker for AWS](aws.md) 
 * [Docker for Azure](azure.md)
 * [Deploying your Apps](deploy.md)

<p style="margin-bottom:50px">&nbsp;</p>


## Getting help

If you need help or would like to discuss setup, 2 forums are available:

* For AWS Help, use the [AWS Forum](https://forums.docker.com/c/docker-for-aws)
* For Azure Help, use the [Azure Forum](https://forums.docker.com/c/docker-for-azure)

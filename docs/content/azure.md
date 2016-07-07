<!--[metadata]>
+++
title = "Docker for Azure"
description = "Docker for Azure"
keywords = ["iaas, aws, azure"]
[menu.iaas]
identifier="docs-azure"
weight="2"
+++
<![end-metadata]-->

# Docker for Azure Setup

## Getting access to the beta

Docker for Azure is currently in private beta. [Sign up](https://beta.docker.com) to get access. When you get into the beta, you will receive an email with an install link and details.

## Prerequisites

- Welcome email
- Access to an Azure account
- SSH key that you want to use when accessing your completed Docker install on Azure

## Configuration

Once you're in the beta, Docker will share with you an Azure Resource Manager template. Deploying the template will orchestrate the deployment of a full Docker swarm with manager and worker nodes.

The simplest way to use the template is with the Azure web portal, but you can also use the CLI or the API.

You'll be prompted for the public key component of an SSH key during setup. You can generate a public key like this:

    ssh-keygen -y -f my-key.pem > pub-key

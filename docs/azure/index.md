<!--[metadata]>
+++
title = "Docker for Azure"
description = "Docker for Azure"
keywords = ["iaas, azure"]
[menu.main]
identifier="docs-azure-index"
parent = "docs-azure"
weight="2"
+++
<![end-metadata]-->

# Docker for Azure Setup

## Getting access to the beta

Docker for Azure is currently in private beta. [Sign up](https://beta.docker.com) to get access. When you get into the beta, you will receive an email with an install link and setup details.

## Prerequisites

- Welcome email
- Access to an Azure account with admin privileges
- SSH key that you want to use when accessing your completed Docker install on Azure

## Configuration

Once you're accepted into the beta, Docker will share with your Azure subscription VHD images required to run Docker. We'll also send you an email with a "Deploy to Azure" button. You can either click the button to deploy Docker for Azure through the Azure web portal or use the url for the Azure Resource Manager (ARM) template (also incluced in the email) to deploy with the CLI.

### Service Principal

To setup Docker for Azure, a [Service Principal](https://azure.microsoft.com/en-us/documentation/articles/active-directory-application-objects/) is required. Docker for Azure uses the principal to operate Azure APIs as you scale up and down or deploy apps on your swarm. Docker provides a containerized helper-script to help create the Service Principal:

    docker run -ti docker4x/create-sp-azure sp-name
    ...
    Your access credentials =============================
    AD App ID:       <app-id>
    AD App Secret:   <secret>

`sp-name` is the name of the authentication app that the script creates with Azure. The name is not important, simply choose something you'll recognize in the Azure portal.

If the script fails, it's typically because your Azure user account doesn't have sufficient privileges. Contact your Azure administrator.

When setting up the ARM template, you will be prompted for the App ID (a UUID) and the app secret.

### SSH Key

Docker for Azure uses SSH for accessing the Docker swarm once it's deployed. During setup, you will be prompted for a SSH public key. If you don't have a SSH key, you can generate one with `puttygen` or `ssh-keygen`. You only need the public key component to setup Docker for Azure. Here's how to get the public key from a .pem file:

    ssh-keygen -y -f my-key.pem

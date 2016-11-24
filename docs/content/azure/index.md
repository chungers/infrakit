<!--[metadata]>
+++
title = "Docker for Azure"
description = "Docker for Azure"
keywords = ["iaas, azure"]
[menu.main]
identifier="docs-azure-index"
parent = "docs-azure"
name = "Setup & Prerequisites"
weight="2"
+++
<![end-metadata]-->

# Docker for Azure Setup

## Getting access to the beta

Docker for Azure is currently in private beta. [Sign up](https://beta.docker.com) to get access. When you get into the beta, you will receive an email with an install link and setup details.

### Docker for Azure private beta sign-up details

When you fill out the sign-up form, make sure you fill in all of the fields, especially the Azure Subscriber ID (36 alphanumeric value, i.e. xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx). Docker for Azure uses a custom VHD that is currently private, and we need your Azure Subscription ID in order to give your account access to the VHD. If you have more than one Azure Subcription that you use (testing, stage, production, etc), email us <docker-for-iaas@docker.com> after you have filled out the form with the list of additional subscription ID that need access. Make sure you put the primary subscriber ID in the form that you filled out, as it might take time for the other subscription IDd to get added to your profile.

You can find your Azure subscription ID by doing the following.

1. Login to the [Azure Portal](https://portal.azure.com/#blade/Microsoft_Azure_Billing/SubscriptionsBlade).
2. On the left hand side menu, select [Subscriptions](https://portal.azure.com/#blade/Microsoft_Azure_Billing/SubscriptionsBlade)

    <img src="/img/azure/subscription.png">

3. Select the subscription you will be using for testing.
3. Copy the subscription identifier from the right-hand column. If you currently do not have an Azure subscription, you can create one on that page.

## Prerequisites

- Welcome email
- Access to an Azure account with admin privileges
- SSH key that you want to use when accessing your completed Docker install on Azure

## Configuration

Once you're accepted into the beta, Docker will share with your Azure subscription VHD images required to run Docker. We'll also send you an email with a "Deploy to Azure" button. You can either click the button to deploy Docker for Azure through the Azure web portal or use the url for the Azure Resource Manager (ARM) template (also incluced in the email) to deploy with the CLI.

### Service Principal

To set up Docker for Azure, a [Service Principal](https://azure.microsoft.com/en-us/documentation/articles/active-directory-application-objects/) is required. Docker for Azure uses the principal to operate Azure APIs as you scale up and down or deploy apps on your swarm. Docker provides a containerized helper-script to help create the Service Principal:

    docker run -ti docker4x/create-sp-azure sp-name
    ...
    Your access credentials =============================
    AD App ID:       <app-id>
    AD App Secret:   <secret>
    AD Tenant ID:   <tenant-id>

If you have multiple Azure subscriptions, make sure you're creating the Service Principal with subscription ID that you shared with Docker when signing up for the beta.

`sp-name` is the name of the authentication app that the script creates with Azure. The name is not important, simply choose something you'll recognize in the Azure portal.

If the script fails, it's typically because your Azure user account doesn't have sufficient privileges. Contact your Azure administrator.

When setting up the ARM template, you will be prompted for the App ID (a UUID) and the app secret.

### SSH Key

Docker for Azure uses SSH for accessing the Docker swarm once it's deployed. During setup, you will be prompted for a SSH public key. If you don't have a SSH key, you can generate one with `puttygen` or `ssh-keygen`. You only need the public key component to set up Docker for Azure. Here's how to get the public key from a .pem file:

    ssh-keygen -y -f my-key.pem

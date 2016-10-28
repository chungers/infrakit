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

Docker for AWS lets you quickly setup and configure a working Docker 1.12 swarm-mode install on Amazon Web Services and on Azure.

Docker for AWS and Azure are available in private beta for testing. They’re free to use (AWS and Azure will charge for resource use).

Sign up for the beta on [beta.docker.com](https://beta.docker.com/).

## Docker for AWS signup details

When you fill out the sign-up form, make sure you fill in all of the fields, especially the AWS Account Number (12 digit value, i.e. 012345678901). Docker for AWS uses a custom AMI that is currently private, and we need your AWS ID in order to give your account access to the AMI. If you have more than one AWS account that you use (testing, stage, production, etc), email us <docker-for-iaas@docker.com> after you have filled out the form with the list of additional account numbers that need access. Make sure you put the primary account in the form that you filled out, as it might take time for the other account numbers to get added to your profile.

You can find your AWS account ID by doing the following.

1. Login to the [AWS Console](https://console.aws.amazon.com/console/home).
2. Click on the [Support link](https://console.aws.amazon.com/support/home?region=us-east-1#/) in the upper right hand corner of the top navigation menu, and click on "Support Center".

    <img src="/img/aws/aws_support_center_link.png">

3. On the Support Center page, in the upper right hand corner you will find your AWS Account Number.

    <img src="/img/aws/aws_account_number.png">

## Docker for Azure signup details

When you fill out the sign-up form, make sure you fill in all of the fields, especially the Azure Subscriber ID (36 alphanumeric value, i.e. xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx). Docker for Azure uses a custom VHD that is currently private, and we need your Azure Subscription ID in order to give your account access to the VHD. If you have more than one Azure Subcription that you use (testing, stage, production, etc), email us <docker-for-iaas@docker.com> after you have filled out the form with the list of additional subscription ID that need access. Make sure you put the primary subscriber ID in the form that you filled out, as it might take time for the other subscription IDd to get added to your profile.

You can find your Azure subscription ID by doing the following.

1. Login to the [Azure Portal](https://portal.azure.com/#blade/Microsoft_Azure_Billing/SubscriptionsBlade).
2. On the left hand side menu, select [Subscriptions](https://portal.azure.com/#blade/Microsoft_Azure_Billing/SubscriptionsBlade)

    <img src="/img/azure/subscription.png">

3. Select the subscription you will be using for testing.
3. Copy the subscription identifier from the right-hand column. If you currently do not have an Azure subscription, you can create one on that page.


## What to know before installing

When setting up Docker for AWS or Azure, you'll be prompted to select manager and worker counts and instance sizes. If you're testing, 1 manager and `small` instances are fine. Both worker count and worker and manager instance size can be changed later, but manager count should not be modified.

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

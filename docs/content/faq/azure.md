<!--[metadata]>
+++
title = "Docker for Azure FAQ"
description = "Docker for Azure FAQ"
keywords = ["iaas, azure, faq"]
[menu.main]
identifier="faq-azure"
parent = "docs-aws-azure-faq"
name = "Azure"
weight="140"
+++
<![end-metadata]-->

# Docker for Azure FAQ

## Why do you need my Azure Subscription ID?
We are using a private Custom AMI, and in order to give you access to this AMI, we need your Azure Subscription ID.

## How do I find my Azure Subscription ID?
You can find this information your Azure Portal Subscription. For more info, look at the directions on [this page](../index.md).

## I use more than one Azure Subscription ID, how do I get access to all of them.
Use the beta sign up form, and put the subscription ID that you need to use most there. Then email us <docker-for-iaas@docker.com> with your information and your other Azure Subscription ID, and we will do our best to add those accounts as well. But due to the large amount of requests, it might take a while before those subscriptions to get added, so be sure to include the important one in the sign up form, so at least you will have that one.

## Can I use my own VHD?
No, at this time we only support our VHD.

## Can I use my existing VM?
Not at this time, but it is on our roadmap for future releases.

## Can I specify the type of Storage Account I use for my VM instances?
Not at this time, but it is on our roadmap for future releases.

## Which Azure regions will this work with.
Docker for Azure should work with all supported marketplace regions.
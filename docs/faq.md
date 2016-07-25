<!--[metadata]>
+++
title = "FAQ"
description = "Docker for AWS Azure FAQ"
keywords = ["iaas, aws, azure"]
[menu.main]
identifier="docs-aws-azure-faq"
weight="100"
+++
<![end-metadata]-->

# FAQ

## Docker for AWS

### Why do you need my Amazon Account Number?
We are using a private Custom AMI, and in order to give you access to this AMI, we need your Amazon account number.

### How do I find my Amazon Account Number?
You can find this information your Amazon Support Center. For more info, look at the directions on [this page](index.md).

### I use more than one Amazon account, how do I get access to all of them.
Use the beta sign up form, and put the account number that you need to use most there. Then email us <docker-for-iaas@docker.com> with your information and your other Amazon account numbers, and we will do our best to add those accounts as well. But due to the large amount of requests, it might take a while before those accounts get added, so be sure to include the important one in the sign up form, so at least you will have that one.

### Can I use my own AMI?
No, at this time we only support our AMI.

### Can I use my existing VPC?
Not at this time, but it is on our roadmap for future releases.

### Can I specify the type of EBS volume I use for my EC2 instances?
Not at this time, but it is on our roadmap for future releases.

### Which AWS regions will this work with.
Docker for AWS should work with all regions except for AWS China, which is a little different than the other regions.

### How many Availability Zones does Docker for AWS use?
All of Amazons regions have at least 2 AZ's, and some have more. To make sure we work in all regions, we currently only support 2 AZ's even if there are more available.

### How long will it take before I get accepted into the private beta?
Docker for AWS is built on top of Docker 1.12 which hasn't officially been released yet, and because of this, things are still changing, which means things can break between release candidates.

We are currently rolling it out slowly to make sure everything is working as it should. This is to ensure that if there are any issues we limit the number of people that are affected.

Once Docker 1.12 goes GA, and we feel there are no major issues that could cause problems, we will start opening it up to more and more people. But please be aware, due to the large number of people on the backlog, it might take a little while to get to you.

### How stable is Docker for AWS
We feel it is fairly stable for development and testing, but since things are consistently changing, we currently don't recommend using it for production workloads at this time.

### I have a suggestion where do I send it?
Send an email to <docker-for-iaas@docker.com> or use the [Docker for AWS Forum](https://forums.docker.com/c/docker-for-aws).

### I have a problem/bug where do I report it?
Send an email to <docker-for-iaas@docker.com> or use the [Dorker for AWS Forum](https://forums.docker.com/c/docker-for-aws)

## Analytics

The beta versions of Docker for AWS and Azure send anonymized analytics to Docker. These analytics are used to monitor beta adoption and are critical to improve Docker for AWS and Azure.

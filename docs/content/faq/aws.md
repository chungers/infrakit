<!--[metadata]>
+++
title = "Docker for AWS FAQ"
description = "Docker for AWS FAQ"
keywords = ["iaas, aws, faq"]
[menu.main]
identifier="faq-aws"
parent = "docs-aws-azure-faq"
name = "AWS"
weight="130"
+++
<![end-metadata]-->

# Docker for AWS FAQ

## Why do you need my Amazon Account Number?
We are using a private Custom AMI, and in order to give you access to this AMI, we need your Amazon account number.

## How do I find my Amazon Account Number?
You can find this information your Amazon Support Center. For more info, look at the directions on [this page](../index.md).

## I use more than one Amazon account, how do I get access to all of them.
Use the beta sign up form, and put the account number that you need to use most there. Then email us <docker-for-iaas@docker.com> with your information and your other Amazon account numbers, and we will do our best to add those accounts as well. But due to the large amount of requests, it might take a while before those accounts to get added, so be sure to include the important one in the sign up form, so at least you will have that one.

## Can I use my own AMI?
No, at this time we only support our AMI.

## Can I use my existing VPC?
Not at this time, but it is on our roadmap for future releases.

## Can I specify the type of EBS volume I use for my EC2 instances?
Not at this time, but it is on our roadmap for future releases.

## Which AWS regions will this work with.
Docker for AWS should work with all regions except for AWS China, which is a little different than the other regions.

## How many Availability Zones does Docker for AWS use?
All of Amazons regions have at least 2 AZ's, and some have more. To make sure we work in all regions, we currently only support 2 AZ's even if there are more available.


## What do I do if I get "KeyPair error" on AWS?
As part of the prerequisites, you need to have an SSH key uploaded to the AWS region you are trying to deploy to. 
For more information about adding an SSH key pair to your account, please refer to the [Amazon EC2 Key Pairs docs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html) 
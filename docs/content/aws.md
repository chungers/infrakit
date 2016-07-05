<!--[metadata]>
+++
title = "Docker for AWS"
description = "Docker for AWS"
keywords = ["iaas, aws, azure"]
[menu.iaas]
identifier="docs-aws"
weight="2"
+++
<![end-metadata]-->

# Docker for AWS Setup

## Getting access to the beta

Docker for AWS is currently in private beta. [Sign up](https://beta.docker.com) to get access. When you get into the beta, you will receive an email with an install link and details.

When you fill out the sign-up form, make sure you fill in all of the fields, especially the AWS Account Number (12 digit value, i.e. 012345678901). Docker for AWS uses a custom AMI that is currently private, and we need your AWS ID in order to give your account access to the AMI. If you have more than one AWS account that you use (testing, stage, production, etc), email us  <docker-for-iaas@docker.com> after you have filled out the form with the list of additional account numbers you need access too. Make sure you put the account in the form that you , as it might take time for the other account numbers to get added to your profile.

You can find your AWS account ID by doing the following.

1. Login to the [AWS Console](https://console.aws.amazon.com/console/home).
2. Click on the [Support link](https://console.aws.amazon.com/support/home?region=us-east-1#/) in the upper right hand corner of the top navigation menu, and click on "Support Center".

    <img src="/img/aws/aws_support_center_link.png">

3. On the Support Center page, in the upper right hand corner you will find your AWS Account Number.

    <img src="/img/aws/aws_account_number.png">

## Prerequisites

- Welcome email
- Access to an AWS account with permissions to use CloudFormation and creating the following objects
    - EC2 instances + Autoscaling groups
    - IAM profiles
    - DynamoDB Tables
    - SQS Queue
    - VPC + subnets
    - ELB
- SSH key in AWS in the region where you want to deploy (required to access the completed Docker install)

## Configuration

Once you're in the beta, Docker will share with you a set of AMIs. You will also get a link to a CloudFormation template that orchestrates deploying a swarm using the AMIs.

The simplest way to use the template is with the AWS web console. The welcome email will include a link.

You can also invoke the template from the AWS CLI:

    $ aws cloudformation create-stack --stack-name teststack --template-url <templateurl> --parameters ParameterKey=KeyName,ParameterValue=<keyname> ParameterKey=InstanceType,ParameterValue=t2.micro ParameterKey=ManagerInstanceType,ParameterValue=t2.micro ParameterKey=ClusterSize,ParameterValue=1 --capabilities CAPABILITY_IAM`

To fully automate installs, you can use the [AWS Cloudformation API](http://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/Welcome.html).

## Modifying Docker install on AWS

### Scaling workers

You can scale the worker count using the AWS Node Autoscaling group. Docker will automatically join or remove new instances to the Swarm.

Changing manager count live is _not_ currently supported.

### Upgrading Docker and changing instance sizes

In the AWS Console, find your CloudFormation stack and select "Update stack". Use the CloudFormation template link. This will let you change the input parameters for the template. AWS will summarize the proposed changes, whether that's changing the AMIs to upgrade Docker or to change instance sizes.

Docker will ensure that upgrade and instance size changes are handled with minimal impact to running apps.

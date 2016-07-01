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

## Prerequisites

- Welcome email
- Access to an AWS account
- SSH key in AWS in the region where you want to deploy (required to access the completed Docker install)

## Configuration

Once you're in the beta, Docker will share with you a set of AMIs. You will also get a link to a CloudFormation template that orchestrates deploying a swarm using the AMIs.

The simplest way to use the template is with the AWS web console. The welcome email will include a link.

You can also invoke the template from the AWS CLI:

    $ aws cloudformation create-stack --stack-name friismteststack --template-url <templateurl> --parameters ParameterKey=KeyName,ParameterValue=<keyname> ParameterKey=InstanceType,ParameterValue=t2.micro ParameterKey=ManagerInstanceType,ParameterValue=t2.micro ParameterKey=ClusterSize,ParameterValue=1 --capabilities CAPABILITY_IAM`

To fully automate installs, you can use the [AWS Cloudformation API](http://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/Welcome.html).

## Modifying Docker install on AWS

### Scaling workers

You can scale the worker count using the AWS Node Autocaling group. Docker will automatically join or remove new instances to the Swarm.

Changing manager count live is _not_ currently supported.

### Upgrading Docker and changing instance sizes

In the AWS Console, find your Cloudformation stack and select "Update stack". Use the Cloudformation template link. This will let you change the input parameters for the template. AWS will summarize the proposed changes, whether that's changing the AMIs to upgrade Docker or to change instance sizes.

Docker will ensure that upgrade and instance size changes are handled with minimal impact to running apps.
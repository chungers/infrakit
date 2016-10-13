<!--[metadata]>
+++
title = "Docker for AWS"
description = "Docker for AWS"
keywords = ["iaas, aws, azure"]
[menu.main]
identifier="docs-aws-index"
parent = "docs-aws"
name = "Setup & Prerequisites"
weight="100"
+++
<![end-metadata]-->

# Docker for AWS Setup

## Getting access to the beta

Docker for AWS is currently in private beta. [Sign up](https://beta.docker.com) to get access. When you get into the beta, you will receive an email with an install link and setup details.

## Prerequisites

- Welcome email
- Access to an AWS account with permissions to use CloudFormation and creating the following objects
    - EC2 instances + Autoscaling groups
    - IAM profiles
    - DynamoDB Tables
    - SQS Queue
    - VPC + subnets
    - ELB
    - CloudWatch Log Group
- SSH key in AWS in the region where you want to deploy (required to access the completed Docker install)

For more information about adding an SSH key pair to your account, please refer to the [Amazon EC2 Key Pairs docs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)


## Configuration

Once you're in the beta, Docker will share with you a set of AMIs. You will also get a link to a CloudFormation template that orchestrates deploying a swarm using the AMIs.

The simplest way to use the template is with the AWS web console. The welcome email will include a link.

You can also invoke the template from the AWS CLI:
```
    $ aws cloudformation create-stack --stack-name teststack --template-url <templateurl> --parameters ParameterKey=KeyName,ParameterValue=<keyname> ParameterKey=InstanceType,ParameterValue=t2.micro ParameterKey=ManagerInstanceType,ParameterValue=t2.micro ParameterKey=ClusterSize,ParameterValue=1 --capabilities CAPABILITY_IAM
```
To fully automate installs, you can use the [AWS Cloudformation API](http://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/Welcome.html).

## How it works
Docker for AWS starts with a CloudFormation template that will create everything that you need from scratch. There are only a few Prerequisites that are listed above.

It first starts off by creating a new VPC along with it's subnets and security groups. Once the networking is setup, it will create two Auto scaling groups, one for the managers and one for the workers, and set the desired capacity that was selected in the CloudFormation setup form. The Managers will start up first and create a Swarm manager quorum using Raft. The workers will then start up and join the swarm one by one, until all of the workers are up and running. At this point you will have x number of managers and y number of workers in your swarm, that are ready to handle your application deployments. See the [deployment](../deploy.md) docs for your next steps.

If you increase the number of instances running in your worker auto scaling group (via the AWS console, or updating the CloudFormation configuration), the new nodes that will start up will automatically join the swarm.

Elastic Load Balancers (ELBs) are setup to help with routing traffic to your swarm.

## System containers
Each node will have a few system containers running on them to help run your swarm cluster. In order for everything to run smoothly, please keep those containers running, and don't make any changes. If you make any changes, we can't guarantee that Docker for AWS will work correctly.

## AMIs
Docker for AWS currently only supports our custom AMI, which is a highly optimized AMI built specifically for running Docker on AWS. These AMI's are not currently public, and in order to use them, we need to give you access to them. As we roll out new AMI's your account will automatically get access to these new versions.

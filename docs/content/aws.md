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

## Docker for AWS Upgrades
There is currently limited support for upgrades, we will be improving this in future releases. The way upgrades work is as follows. We will release a new CloudFormation template, and you will update your CloudFormation stack with the new CloudFormation template. CloudFormation will look at your current stack, and see what is different, and let you know what it is about to do. If you confirm the changes it will start to update your stack.

If there is a change to the AMI, then one by one the old nodes will be shut down, and replaced by newer nodes. How long this takes will depend on how many nodes you have. We do a slow rolling update, so the more nodes, the longer it will take. Eventually when complete, all of the older nodes will be gone, and replaced with a completely new set of nodes.

Since we are doing a slow rolling upgrade, the services that are running on a node that is getting shut down, will be rescheduled by swarm, and put on to a new healthy node. If your service is properly scaled it should not notice any downtime.

Since this feature is still very new, there is a chance things could go wrong, so we don't recommend using Docker for AWS for production critical workloads at this time.

## How it works
Docker for AWS starts with a CloudFormation template that will create everything that you need from scratch. There are only a few Prerequisites that are listed above.

It first starts off by creating a new VPC along with it's subnets and security groups. Once the networking is setup, it will create two Auto scaling groups, one for the managers and one for the workers, and set the desired capacity that was selected in the CloudFormation setup form. The Managers will start up first and create a Swarm manager quorum using Raft. The workers will then start up and join the swarm one by one, until all of the workers are up and running. At this point you will have x number of managers and y number of workers in your swarm, that are ready to handle your application deployments. See the [deployment](deploy.md) docs for your next steps.

If you increase the number of instances running in your worker auto scaling group (via the AWS console, or updating the CloudFormation configuration), the new nodes that will start up will automatically join the swarm.

Elastic Load Balancers (ELBs) are setup to help with routing traffic to your swarm.

## System containers
Each node will have a few system containers running on them to help run your swarm cluster. In order for everything to run smoothly, please keep those containers running, and don't make any changes. If you make any changes, we can't guarantee that Docker for AWS will work correctly.

## AMIs
Docker for AWS currently only supports our custom AMI, which is a highly optimized AMI built specifically for running Docker on AWS. These AMI's are not currently public, and in order to use them, we need to give you access to them. As we roll out new AMI's your account will automatically get access to these new versions.

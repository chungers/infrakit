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

## Prerequisites

- Access to an AWS account with permissions to use CloudFormation and creating the following objects
    - EC2 instances + Auto Scaling groups
    - IAM profiles
    - DynamoDB Tables
    - SQS Queue
    - VPC + subnets
    - ELB
    - CloudWatch Log Group
- SSH key in AWS in the region where you want to deploy (required to access the completed Docker install)
- AWS account that support EC2-VPC [For more info about EC2-Classic](../faq/aws.md)

For more information about adding an SSH key pair to your account, please refer to the [Amazon EC2 Key Pairs docs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)


## Configuration

Docker for AWS has a CloudFormation template that orchestrates deploying a swarm using our custom AMIs. There are two ways you can deploy Docker for AWS. You can use the AWS Management Console (browser based), or use the AWS CLI. Both have the following configuration options.

### Configuration options

#### KeyName
Pick the SSH key that will be used when you SSH into the manager nodes.

#### InstanceType
The EC2 instance type for your worker nodes

#### ManagerInstanceType
The EC2 instance type for your manager nodes. The larger your swarm, the larger the instance size you should use.

#### ClusterSize
The number of workers you want in your swarm? (1-1000)

#### ManagerSize
The number of Managers in your swarm. You can pick either 1, 3 or 5 managers. We only recommend 1 manager for testing and dev setups. There are no failover guarantee's with 1 manager, if it fails the swarm will go down as well.

We recommend at least 3 managers, and if you have a lot of workers, you should pick 5 managers.

#### EnableSystemPrune
Enable if you want Docker for AWS to automatically cleanup unused space on your swarm nodes

Every day at 1:42AM UTC, it will run `docker system prune` on each of your nodes (workers, and managers). Each run is staggered slightly so that they are not run on all nodes at the same time. This limits any resource spikes on the swarm that resource cleanup might cause.

This will remove the following:
- All stopped containers
- All volumes not used by at least one container
- All dangling images
- All unused networks

#### WorkerDiskSize
Size of Workers's ephemeral storage volume in GiB (20 - 1024)

#### WorkerDiskType
Worker ephemeral storage volume type ("standard", "gp2")

#### ManagerDiskSize
Size of Manager's ephemeral storage volume in GiB (20 - 1024)

#### ManagerDiskType
Manager ephemeral storage volume type ("standard", "gp2")

### Installing with the AWS Management Console
The simplest way to use the template is with the CloudFormation section of the AWS Management Console.

Go to the [Release notes](../deploy.md) page, and click on the "launch stack" button, to start the deployment process.

### Installing with the CLI
You can also invoke the template from the AWS CLI:

Here is an example of how to use the CLI, make sure you populate all of the parameters and their values.
```
$ aws cloudformation create-stack --stack-name teststack --template-url <templateurl> --parameters ParameterKey=KeyName,ParameterValue=<keyname> ParameterKey=InstanceType,ParameterValue=t2.micro ParameterKey=ManagerInstanceType,ParameterValue=t2.micro ParameterKey=ClusterSize,ParameterValue=1 --capabilities CAPABILITY_IAM
```

To fully automate installs, you can use the [AWS Cloudformation API](http://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/Welcome.html).

## How it works
Docker for AWS starts with a CloudFormation template that will create everything that you need from scratch. There are only a few Prerequisites that are listed above.

It first starts off by creating a new VPC along with its subnets and security groups. Once the networking is set up, it will create two Auto Scaling groups, one for the managers and one for the workers, and set the desired capacity that was selected in the CloudFormation setup form. The Managers will start up first and create a Swarm manager quorum using Raft. The workers will then start up and join the swarm one by one, until all of the workers are up and running. At this point you will have x number of managers and y number of workers in your swarm, that are ready to handle your application deployments. See the [deployment](../deploy.md) docs for your next steps.

If you increase the number of instances running in your worker Auto Scaling group (via the AWS console, or updating the CloudFormation configuration), the new nodes that will start up will automatically join the swarm.

Elastic Load Balancers (ELBs) are set up to help with routing traffic to your swarm.

## System containers
Each node will have a few system containers running on them to help run your swarm cluster. In order for everything to run smoothly, please keep those containers running, and don't make any changes. If you make any changes, we can't guarantee that Docker for AWS will work correctly.

## AMIs
Docker for AWS currently only supports our custom AMI, which is a highly optimized AMI built specifically for running Docker on AWS. These AMI's are not currently public, and in order to use them, we need to give you access to them. As we roll out new AMI's your account will automatically get access to these new versions.

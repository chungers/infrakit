# Docker for AWS

## Prerequisites
- Access to an AWS account. To get access to the Docker for AWS AMIs, ping us in the #editions Slack channel
    - To find your AWS account ID number in the AWS Management Console, click on Support in the navigation bar in the upper-right, and then click Support Center. Your currently signed in account ID appears below the Support menu.
- AWS Console or CLI access with permissions to run CloudFormation templates
- An SSH key in AWS in the region where you want to deploy (required to SSH into machines and run docker/swarm commands)
- Logged into the AWS console
- Make sure that you have enough capacity for the swarm that you want to build, and won't go over any of your limits.

Once you have all of the above you are ready to move onto the next step.

## Configuration

You can use the AWS CLI to invoke the template.  e.g.:

    $ aws cloudformation create-stack \
        --stack-name friismteststack \
        --capabilities CAPABILITY_IAM \
        --template-url https://docker-for-aws.s3.amazonaws.com/aws/beta/aws-v1.12.1-beta5.json \
        --parameters ParameterKey=KeyName,ParameterValue=friism-us-west-1 \
        ParameterKey=InstanceType,ParameterValue=t2.micro \
        ParameterKey=ManagerInstanceType,ParameterValue=t2.micro \
        ParameterKey=ClusterSize,ParameterValue=1

Note: The `Makefile` in this directory in the repository can invoke this for you
automatically if you set a few variables such as `KEYNAME` for SSH key.  e.g.:

    $ make USER=$(whoami) KEYNAME=uploaded-key

It can also tear down created stack(s) via `make clean`, e.g.:

    $ make USER=$(whoami) clean

A local development version of Docker for AWS can be invoked via `make` using
the CloudFormation template stored in your local repository instead of a remote
one (the default used by `make`).  This is useful for developing the
CloudFormation JSON itself.  Simply set `DEV=1`.

    $ make USER=$(whoami) KEYNAME=mykey DEV=1

... or click this button:

[![Docker for AWS](https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home?#/stacks/new?stackName=Docker&templateURL=https://docker-for-aws.s3.amazonaws.com/aws/beta/aws-v1.12.1-beta5.json)

This will take you to the AWS console and preload the CloudFormation template. Hit "Next" on the first prompt.

You'll be asked how many managers and workers you want to run. Decide how much resilience you need, then click the "Launch Stack" button to launch your stack.

| # of managers  | # of tolerated failures |
| ------------- | ------------- |
| 1  | 0  |
| 3  | 1  |
| 5  | 2  |

This will create the stack in the background. When complete, you can go to the "outputs" tab in the CloudFormation stack list detail page. The outputs will show how to SSH to an ELB host. Your SSH session will be on one of the manager nodes.

Once you are logged into the container you can run Docker commands on the cluster:

    $ docker swarm info
    $ docker node ls

You can also tunnel the Docker socket over SSH to remotely run commands on the cluster (replace `managerPublicIP` with your ELB hostname):

    $ ssh -NL localhost:2375:/var/run/docker.sock docker@managerPublicIP &
    $ docker -H localhost:2375 info

## Running apps

You can now start creating containers and services.

    $ docker run hello-world

To run websites, you currently have to apply a label to let the edition now to expose a port through the ELB:

    $ docker network create -d overlay foo
    $ docker service create --name nginx --scale 1 --network foo -p 80:80/tcp --label docker.swarm.lb.port=80/http nginx

Once up, you access the site on the ELB hostname.

You can also run the voting app:

    $ docker network create -d overlay mynetwork
    $ docker service create --scale 1 --name redis --network mynetwork redis:latest
    $ docker service create --scale 1 --name pg --env POSTGRES_PASSWORD=pg8675309 --network mynetwork postgres:latest
    $ docker service create --scale 1 --name frontend --env QUEUE_HOSTNAME=redis --env OPTION_A=Kats --env OPTION_B=Doggies --network mynetwork -p 80:80/tcp --label docker.swarm.lb.port=80/http mikegoelzer/s2_frontend:latest
    $ docker service create --scale 1 --name worker --network mynetwork mikegoelzer/s2_worker:latest
    $ docker service create --scale 1 --name results --env PORT=8080 --network mynetwork --label docker.swarm.lb.port=81/http -p 8080:8080 mikegoelzer/s2_results:latest


## Cleanup

When you are done with the swarm, please remember to destroy it. You can do that from the CloudFormation stack list by selecting "Delete Stack".




# AWS release script
This script will take some parameters and then it will do the following.

1. copy the AMI to the different regions
2. share the image across the list of accounts
3. generate a new CloudFormation template and upload to the `editions-stage-us-east-1-150610-005505` s3 bucket
4. return the s3 url for the CloudFormation template, to be used in release notes, etc.


## prerequisites
1. Access to the distribution AWS account, and an IAM account with permissions on the `editions-stage-us-east-1-150610-005505` s3 bucket.
2. Set Env variables for account referenced in 1.
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
3. Make sure the aws accountId file is up to date in s3. (https://s3.amazonaws.com/docker-for-aws/data/accounts.txt)


## Running release script.

1. Checkout the master branch from `https://github.com/docker/editions` and make sure local copy is up to date.
2. Calling `make release` script, passing in correct variables. For example:

    $ make release EDTIONS_COMMIT= RELEASE=

    $ make release EDTIONS_COMMIT=1234567asdfghjkl RELEASE=true

    EDTIONS_COMMIT will determine what AMI will be copied and what version of Docker will be published

3. Look at the logs and find the CloudFormation s3 url, and give that to who needs it.
4. Tag the code and push tags to github, and create release from tag.


### Development Notes
Here are notes that were used during development, some are still relevant so keeping here for now.

Ideal setup:

1. Given a git hash of editions download the AMI built and read metadata to get Docker version.
2. Share that AMI across regions, return a dict of AMI IDs, store in s3
   <EDITIONS_COMMIT>/ami_list.json
   /ami_list.json
3. Given the ami_list.json and an account_list.txt approve the AMI to the accounts.

GET AMI_ID >> COPY_AMI >> LIST_OF_AMIS + List of Accounts >> PUBLISH_AMIS >> generate CFN >> upload CFN to S3

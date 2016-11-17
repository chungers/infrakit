# AWS release script
This script will take some parameters and then it will do the following.

1. copy the AMI to the different regions
2. share the image across the list of accounts
3. generate a new CloudFormation template and upload to the `docker-for-aws` s3 bucket
4. return the s3 url for the CloudFormation template, to be used in release notes, etc.


## prerequisites
1. Access to the distribution AWS account, and an IAM account with permissions on the `docker-for-aws` s3 bucket.
2. Set Env variables for account referenced in 1.
- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
3. Make sure the aws accountId file is up to date in s3. (https://s3.amazonaws.com/docker-for-aws/data/accounts.txt)


## Running release script.

1. Checkout the master branch from `https://github.com/docker/editions` and make sure local copy is up to date.
2. Calling `aws/release/run_release.sh` script, passing in correct variables. For example:

    $ aws/release/run_release.sh -d docker_version -e edition_version -a ami_id -r ami_src_region -p no

    $ aws/release/run_release.sh -d "1.12.0" -e "beta4" -a "ami-148c4e74" -r "us-west-2" -p no

    optional fields:
    -c channel (defaults to beta) [beta, nightly, alpha, etc]
    -l AWS account list url, url for the list of Answers4AWS accounts we want to give access to the AMI.
    -p make the ami public? (yes, no)

3. Look at the logs and find the CloudFormation s3 url, and give that to who needs it.
4. Tag the code and push tags to github, and create release from tag.


## CS / DDC builds
DDC depends on Docker CS. We don't builds a CS binary every night, so there is no need to build AMI's every night as well.
At least not yet, I'm sure there will be reasons in the near future.

Steps:
1. build AMI and update latest + Release AMI (share with accounts, and all regions)
    ./build-cs-ami.sh <URL to docker binary> yes
    example: ./build-cs-ami.sh https://s3.amazonaws.com/packages.docker.com/1.12/builds/linux/amd64/docker-1.12.1-cs1.tgz yes
2. build ddc (creates the CFN template) using AMI's from 1.
    ./run_ddc_release.sh -d "1.12.0-cs1" -e "alpha4" -c "alpha"


### Development Notes
Here are notes that were used during development, some are still relevant so keeping here for now.

Ideal setup:

1. Given a git hash of docker, and moby, and version build a moby AMI.
2. Share that AMI across regions, return a dict of AMI IDs, store in s3
    /release/version/version.txt
                    /amis.json
3. Given the amis.json and an account_list.txt approve the AMI to the accounts.

/release/amis/[nightly|beta/alpha]/

Bake AMI >> AMI_ID >> COPY_AMI >> LIST_OF_AMIS + List of Accounts >> PUBLISH_AMIS >> generate CFN >> upload CFN to S3 >> Email accounts of new release.

Current way they approve accounts:

cat awsid.txt| awk '{print "bash approve-account.sh "$1" ami.txt"}' | bash

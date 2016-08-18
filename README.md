# Docker Editions

### Release Process (Automated)

Today the release process must be invoked within an Amazon EC2 instance.  This
is so the created AMI can be built by mounting in an EBS volume.  If the region
this instance is in is not `us-west-1`, make sure to set `REGION` differently
when invoking `make release` as outlined below.

To run the automated release process:

1. Set `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AZURE_STG_ACCOUNT_KEY`
   environment variables.  The first two are for authenticating your IAM role
   with the AWS API, and `AZURE_STG_ACCOUNT_KEY` is the key for accessing the
   `dockereditions` (you can access this in the account's settings in the Azure
   portal). _(Note): At the time of writing the Moby build process also assumes
   valid credentials stored at `$HOME/.aws/credentials`, but this will be
   revised soon to require only the environment variables)._
2. Revise the `Makefile` top-level variables to have the proper settings, e.g.
   Docker version and `betaX` tag.
3. Run `make clean && make release` (`make release REGION=us-east-1` if needed,
   etc.).  `make release` will build the needed
   [Moby](https://github.com/docker/moby) images (AMI + VHD) and run through the
   build steps outlined in the next section.  You will still need to follow
   through on updating the account list, building the e-mail, testing the template,
   etc.

### Release Process (Manual)

1. Checkout latest `docker/editions` code
2. Build the `buoy` app.
    - `cd tools/buoy && ./build_buoy.sh && cd -`
3. Run `build_and_push_all.sh` scripts with proper versions.
    - `cd aws/dockerfiles && VERSION=aws-vx.y.z-betaN ./build_and_push_all.sh && cd -`
    - `cd azure/dockerfiles && VERSION=azure-vx.y.z-betaN ./build_and_push_all.sh && cd -`
4. Add your AWS ENV variables for the AWS distribution account.
5. Build the AWS AMI, get the ami-id and region from the build.
6. Update the aws accounts txt doc (https://s3.amazonaws.com/docker-for-aws/data/accounts.txt)
    - Source : https://drive.google.com/a/docker.com/file/d/0Bw5EwLyGjFkQNXpCY0tzWWZlNHc/view?ts=57aa1406
    - Ask French ben for the file if you don't have access to google drive file
    - make sure the file is made public.
7. Run the release script: (change variables to match the release)
    - cd aws/release/
    - example: ./run_release.sh -d 1.12.1-rc1 -e beta5 -a ami-61f6b501 -r us-west-1 -c beta
8. Get the URL output from release script, give to person creating release notes and sending emails.
9. Test out CloudFormation template before sending out emails.
10. Update beta docs site
11. Send out emails

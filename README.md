# editions

## Docker for AWS
Docker for AWS provides an easy way for someone to go from nothing to a swarm cluster in only a few clicks.

Located under the `/aws` directory.

### Release
1. checkout latest editions code
2. update the VERSION variables in aws/azure build_and_push_all.sh
    - aws/dockerfiles/build_and_push_all.sh
    - azure/dockerfiles/build_and_push_all.sh
2. build the buoy app
    tools/buoy/build_buoy.sh
3. Build images for aws/azure
    - aws/dockerfiles/build_and_push_all.sh
    - azure/dockerfiles/build_and_push_all.sh
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


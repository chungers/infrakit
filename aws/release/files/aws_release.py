#!/usr/bin/env python
import json
import argparse

from utils import (
    get_ami, wait_for_ami_to_be_available, copy_amis, strip_and_padd, get_account_list,
    approve_accounts, ACCOUNT_LIST_FILE_URL, create_cfn_template)


CFN_TEMPLATE = '/home/docker/docker_for_aws.template'
CFN_CLOUD_TEMPLATE = '/home/docker/docker_for_aws_cloud.template'

def main():
    parser = argparse.ArgumentParser(description='Release Docker for AWS')
    parser.add_argument('-d', '--docker_version',
                        dest='docker_version', required=True,
                        help="Docker version (i.e. 1.12.0-rc4)")
    parser.add_argument('-e', '--edition_version',
                        dest='edition_version', required=True,
                        help="Edition version (i.e. Beta 4)")
    parser.add_argument('-a', '--ami_id',
                        dest='ami_id', required=True,
                        help="ami-id for the Moby AMI we just built (i.e. ami-123456)")
    parser.add_argument('-s', '--ami_src_region',
                        dest='ami_src_region', required=True,
                        help="The reason the source AMI was built in (i.e. us-east-1)")
    parser.add_argument('-c', '--channel',
                        dest='channel', default="beta",
                        help="release channel (beta, alpha, rc, nightly)")
    parser.add_argument('-u', '--channel_cloud',
                        dest='channel_cloud', default="alpha",
                        help="DDC release channel (beta, alpha, rc, nightly)")
    parser.add_argument('-l', '--account_list_url',
                        dest='account_list_url', default=ACCOUNT_LIST_FILE_URL,
                        help="The URL for the aws account list for ami approvals")

    args = parser.parse_args()

    release_channel = args.channel
    release_cloud_channel = args.channel_cloud
    docker_version = args.docker_version
    # TODO change to something else? where to get moby version?
    moby_version = docker_version
    edition_version = args.edition_version
    flat_edition_version = edition_version.replace(" ", "")
    docker_for_aws_version = u"aws-v{}-{}".format(docker_version, flat_edition_version)
    image_name = u"Moby Linux {}".format(docker_for_aws_version)
    image_description = u"The best OS for running Docker, version {}".format(moby_version)
    print("\n Variables")
    print(u"release_channel={}".format(release_channel))
    print(u"docker_version={}".format(docker_version))
    print(u"edition_version={}".format(edition_version))
    print(u"ami_id={}".format(args.ami_id))
    print(u"ami_src_region={}".format(args.ami_src_region))
    print(u"account_list_url={}".format(args.account_list_url))
    if not args.account_list_url:
        print("account_list_url parameter is None, defaulting")
        account_list_url = ACCOUNT_LIST_FILE_URL
    else:
        account_list_url = args.account_list_url
    print(u"account_list_url={}".format(account_list_url))

    print("Copy AMI to each region..")
    ami_list = copy_amis(args.ami_id, args.ami_src_region,
                         image_name, image_description, release_channel)
    ami_list_json = json.dumps(ami_list, indent=4, sort_keys=True)
    print(u"AMI copy complete. AMI List: \n{}".format(ami_list_json))
    print("Get account list..")
    account_list = get_account_list(account_list_url)
    print(u"Approving AMIs for {} accounts..".format(len(account_list)))
    approve_accounts(ami_list, account_list)
    print("Accounts have been approved.")

    print("Create CloudFormation template..")
    s3_url = create_cfn_template(ami_list, release_channel, docker_version,
                                 docker_for_aws_version, edition_version, CFN_TEMPLATE,
                                 docker_for_aws_version)
    s3_cloud_url = create_cfn_template(ami_list, release_cloud_channel, docker_version,
                                 docker_for_aws_version, edition_version, CFN_CLOUD_TEMPLATE,
                                 docker_for_aws_version)

    # TODO: git commit, tag release. requires github keys, etc.

    print("------------------")
    print(u"Finshed.. CloudFormation URL={0} \n\t cloud-URL={1} \n".format(s3_url, s3_cloud_url))
    print("Don't forget to tag the code (git tag -a v{0}-{1} -m {0}; git push --tags)".format(
        docker_version, flat_edition_version))
    print("------------------")


if __name__ == '__main__':
    main()

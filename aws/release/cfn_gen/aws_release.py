#!/usr/bin/env python
import json
import argparse

from utils import (
    copy_amis, get_account_list, set_ami_public,
    approve_accounts, ACCOUNT_LIST_FILE_URL, create_cfn_template)

from base import AWSBaseTemplate
from cloud import CloudVPCTemplate, CloudVPCExistingTemplate
from existing_vpc import ExistingVPCTemplate
from docker_ee import DockerEEVPCTemplate, DockerEEVPCExistingTemplate


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
    parser.add_argument('-p', '--public',
                        dest='make_ami_public', default="no",
                        help="Make the AMI public")
    parser.add_argument('-o', '--edition_addon',
                        dest='edition_addon', default="base",
                        help="Edition Add-On (base, ddc, cloud, etc.)")

    args = parser.parse_args()

    release_channel = args.channel
    release_cloud_channel = args.channel_cloud
    docker_version = args.docker_version
    # TODO change to something else? where to get moby version?
    moby_version = docker_version
    edition_version = args.edition_version
    edition_addon = args.edition_addon
    flat_edition_version = edition_version.replace(" ", "")
    docker_for_aws_version = u"aws-v{}-{}".format(docker_version, flat_edition_version)
    image_name = u"Moby Linux {}".format(docker_for_aws_version)
    image_description = u"The best OS for running Docker, version {}".format(moby_version)
    print("\n Variables")
    print(u"release_channel={}".format(release_channel))
    print(u"docker_version={}".format(docker_version))
    print(u"edition_version={}".format(edition_version))
    print(u"edition_addon={}".format(edition_addon))
    print(u"ami_id={}".format(args.ami_id))
    print(u"ami_src_region={}".format(args.ami_src_region))
    print(u"account_list_url={}".format(args.account_list_url))
    print(u"make_ami_public={}".format(args.make_ami_public))
    if not args.account_list_url:
        print("account_list_url parameter is None, defaulting")
        account_list_url = ACCOUNT_LIST_FILE_URL
    else:
        account_list_url = args.account_list_url
    print(u"account_list_url={}".format(account_list_url))

    if not args.make_ami_public:
        make_ami_public = False
    else:
        make_ami_public = args.make_ami_public
        if make_ami_public.lower() == 'yes':
            make_ami_public = True
        else:
            make_ami_public = False
    print(u"make_ami_public={}".format(make_ami_public))

    print("Copy AMI to each region..")
    ami_list = copy_amis(args.ami_id, args.ami_src_region,
                         image_name, image_description, release_channel)
    ami_list_json = json.dumps(ami_list, indent=4, sort_keys=True)
    print(u"AMI copy complete. AMI List: \n{}".format(ami_list_json))

    if make_ami_public:
        print("Make AMI's public.")
        # we want to make this public.
        set_ami_public(ami_list)
        print("Finished making AMI's public.")
    else:
        print("Get account list..")
        account_list = get_account_list(account_list_url)
        print(u"Approving AMIs for {} accounts..".format(len(account_list)))
        approve_accounts(ami_list, account_list)
    print("Accounts have been approved.")

    print("Create CloudFormation template..")
    cfn_name = docker_for_aws_version
    s3_url = create_cfn_template(AWSBaseTemplate, ami_list, release_channel,
                                 docker_version, edition_version,
                                 docker_for_aws_version, edition_addon, cfn_name)

    cfn_name = "{}-no-vpc".format(docker_for_aws_version)
    s3_url_no_vpc = create_cfn_template(ExistingVPCTemplate, ami_list,
                                        release_channel,
                                        docker_version, edition_version,
                                        docker_for_aws_version, edition_addon, cfn_name,
                                        cfn_type="no-vpc")

    docker_ee_cfn_name = docker_for_aws_version
    docker_ee_release_channel = u"{}-ee".format(release_channel)
    docker_ee_s3_url = create_cfn_template(DockerEEVPCTemplate, ami_list,
                                           docker_ee_release_channel,
                                           docker_version, edition_version,
                                           docker_for_aws_version,
                                           docker_ee_cfn_name)

    docker_ee_cfn_name = "{}-no-vpc".format(docker_ee_cfn_name)
    docker_ee_s3_url_no_vpc = create_cfn_template(DockerEEVPCExistingTemplate,
                                                  ami_list,
                                                  docker_ee_release_channel,
                                                  docker_version,
                                                  edition_version,
                                                  docker_for_aws_version,
                                                  docker_ee_cfn_name,
                                                  cfn_type="no-vpc")

    cfn_name = "{}-cloud".format(docker_for_aws_version)
    s3_cloud_url = create_cfn_template(CloudVPCTemplate, ami_list,
                                       release_cloud_channel,
                                       docker_version, edition_version,
                                       docker_for_aws_version, edition_addon, cfn_name)

    cfn_name = "{}-no-vpc-cloud".format(docker_for_aws_version)
    s3_cloud_url_no_vpc = create_cfn_template(CloudVPCExistingTemplate,
                                              ami_list, release_cloud_channel,
                                              docker_version, edition_version,
                                              docker_for_aws_version, edition_addon,
                                              cfn_name, cfn_type="no-vpc")

    # TODO: git commit, tag release. requires github keys, etc.

    print("------------------")
    print(u"Finshed.. Docker CE Base URL={}".format(s3_url))
    print(u"Finshed.. Docker CE Base No VPC URL={}".format(s3_url_no_vpc))
    print(u"Finshed.. Docker EE Base URL={}".format(docker_ee_s3_url))
    print(u"Finshed.. Docker EE Base No VPC URL={}".format(docker_ee_s3_url_no_vpc))
    print(u"Finshed.. CloudFormation Cloud URL={}".format(s3_cloud_url))
    print(u"Finshed.. CloudFormation Cloud No VPC URL={}".format(s3_cloud_url_no_vpc))
    print("Don't forget to tag the code (git tag -a v{0}-{1} -m {0}; git push --tags)".format(
        docker_version, flat_edition_version))
    print("------------------")


if __name__ == '__main__':
    main()

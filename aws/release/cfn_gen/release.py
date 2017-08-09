#!/usr/bin/env python
import json
import argparse

from utils import (
    copy_amis, get_ami_list, get_account_list, set_ami_public, upload_ami_list, generate_ami_list_url, set_ee_ami_list, 
    approve_accounts, ACCOUNT_LIST_FILE_URL, create_cfn_template, upload_cfn_template)

from docker_ce import DockerCEVPCTemplate, DockerCEVPCExistingTemplate
from cloud import CloudVPCTemplate, CloudVPCExistingTemplate
from docker_ee import DockerEEVPCTemplate, DockerEEVPCExistingTemplate
from ddc import DDCVPCExistingTemplate, DDCVPCTemplate
from ddc_dev import DDCDevVPCTemplate, DDCDevVPCExistingTemplate


def main():
    parser = argparse.ArgumentParser(description='Release Docker for AWS')
    parser.add_argument('-d', '--docker_version',
                        dest='docker_version', required=True,
                        help="Docker version (i.e. 17.03.0-ce)")
    parser.add_argument('-e', '--editions_version',
                        dest='editions_version', required=True,
                        help="Edition version (i.e. 17.03.0-ce-aws1)")
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
    parser.add_argument("--upload", action="store_true",
                        help="Upload the AWS template once generated")
    parser.add_argument("--share", 
                        dest='share_ami', default="yes",
                        help="Share the AWS AMI with provided account list")
    parser.add_argument("--enterprise", 
                        dest='is_ee', default="no",
                        help="Build the Enterprise Edition")
    parser.add_argument("--template", action="store_true",
                        help="Generate the AWS template using moby commit AMI list")

    args = parser.parse_args()

    release_channel = args.channel
    release_cloud_channel = args.channel_cloud
    docker_version = args.docker_version
    # TODO change to something else? where to get moby version?
    moby_version = docker_version
    docker_for_aws_version = args.editions_version
    if not docker_for_aws_version:
        raise Exception("No Editions Version was defined")
    edition_addon = args.edition_addon
    flat_editions_version = docker_for_aws_version.replace(" ", "")
    image_name = u"Moby Linux {} {}".format(docker_for_aws_version, release_channel)
    image_description = u"The best OS for running Docker, version {}".format(moby_version)
    ami_id = ''
    if args.ami_id:
        ami_id = args.ami_id
    
    share_ami = ''
    if args.share_ami:
        share_ami = args.share_ami
    
    make_ee = False
    if args.is_ee:
        if args.is_ee.lower() == 'yes':
            make_ee = True
            # don't share AMI if it's EE
            share_ami = "no"
    
    print("\n Variables")
    print(u"release_channel={}".format(release_channel))
    print(u"docker_version={}".format(docker_version))
    print(u"docker_for_aws_version={}".format(docker_for_aws_version))
    print(u"edition_addon={}".format(edition_addon))
    print(u"make_ee={}".format(make_ee))
    if args.template:
        ami_list = get_ami_list(generate_ami_list_url())
    else:
        if not args.account_list_url:
            print("account_list_url parameter is None, defaulting")
            account_list_url = ACCOUNT_LIST_FILE_URL
        else:
            account_list_url = args.account_list_url
        print(u"account_list_url={}".format(account_list_url))
        print(u"ami_id={}".format(args.ami_id))
        print(u"ami_src_region={}".format(args.ami_src_region))
        print(u"make_ami_public={}".format(args.make_ami_public))
        make_ami_public = False
        if args.make_ami_public:
            if args.make_ami_public.lower() == 'yes':
                make_ami_public = True
        print(u"make_ami_public={}".format(make_ami_public))

        if make_ee:
            ami_list = set_ee_ami_list(args.ami_id, args.ami_src_region)
        else:
            print("Copy AMI to each region..")
            ami_list = copy_amis(args.ami_id, args.ami_src_region,
                            image_name, image_description, release_channel)
        print(u"AMI copy complete. AMI List: \n{}".format(ami_list))

        ami_list_json = json.dumps(ami_list, indent=4, sort_keys=True)
        print("Upload AMI list to s3")
        upload_ami_list(ami_list_json, docker_for_aws_version, docker_version, release_channel)

        if make_ami_public:
            print("Make AMI's public.")
            # we want to make this public.
            set_ami_public(ami_list)
            print("Finished making AMI's public.")
        elif share_ami.lower() == 'yes':
            print("Get account list..")
            account_list = get_account_list(account_list_url)
            print(u"Approving AMIs for {} accounts..".format(len(account_list)))
            approve_accounts(ami_list, account_list)
            print("Accounts have been approved.")

    print("Create CloudFormation template..")

    cfn_name = u"Docker"
    base_url = create_cfn_template(DockerCEVPCTemplate, ami_list, release_channel,
                                 docker_version, docker_for_aws_version, edition_addon, cfn_name)

    no_vpc_cfn_name = "{}-no-vpc".format(cfn_name)
    base_url_no_vpc = create_cfn_template(DockerCEVPCExistingTemplate, ami_list, release_channel,
                                        docker_version, docker_for_aws_version,
                                        edition_addon, no_vpc_cfn_name,
                                        cfn_type="no-vpc")

    docker_ee_cfn_name = u"{}-ee".format(cfn_name)
    docker_ee_release_channel = u"{}-ee".format(release_channel)
    docker_ee_base_url = create_cfn_template(DockerEEVPCTemplate, ami_list,
                                           docker_ee_release_channel,
                                           docker_version, docker_for_aws_version, edition_addon,
                                           docker_ee_cfn_name)

    docker_ee_no_vpc_cfn_name = "{}-no-vpc-ee".format(cfn_name)
    docker_ee_base_url_no_vpc = create_cfn_template(DockerEEVPCExistingTemplate,
                                                  ami_list,
                                                  docker_ee_release_channel,
                                                  docker_version, docker_for_aws_version,
                                                  edition_addon, docker_ee_no_vpc_cfn_name,
                                                  cfn_type="no-vpc")

    cloud_cfn_name = "{}-Cloud".format(cfn_name)
    edition_addon = 'cloud'
    cloud_url = create_cfn_template(CloudVPCTemplate, ami_list,
                                       release_cloud_channel,
                                       docker_version, docker_for_aws_version, 
                                       edition_addon, cloud_cfn_name)

    cloud_no_vpc_cfn_name = "{}-Cloud-no-vpc".format(cfn_name)
    cloud_url_no_vpc = create_cfn_template(CloudVPCExistingTemplate,
                                              ami_list, release_cloud_channel,
                                              docker_version, docker_for_aws_version,
                                              edition_addon, cloud_no_vpc_cfn_name, cfn_type="no-vpc")

    # DDC
    ddc_channel = "{}-ddc".format(release_channel)
    ddc_cfn_name = "{}-ddc".format(cfn_name)
    edition_addon = 'ddc'
    ddc_url = create_cfn_template(DDCVPCTemplate, ami_list,
                                     ddc_channel,
                                     docker_version, docker_for_aws_version,
                                     edition_addon, ddc_cfn_name)

    ddc_no_vpc_cfn_name = "{}-ddc-no-vpc".format(cfn_name)
    ddc_url_no_vpc = create_cfn_template(DDCVPCExistingTemplate,
                                            ami_list, ddc_channel,
                                            docker_version, docker_for_aws_version,
                                            edition_addon, ddc_no_vpc_cfn_name, cfn_type="no-vpc")

    # DDC-Dev
    ddc_dev_cfn_name = "{}-ddc-dev".format(cfn_name)
    edition_addon = 'ddc-dev'
    ddc_dev_url = create_cfn_template(DDCDevVPCTemplate, ami_list,
                                         ddc_channel,
                                         docker_version, docker_for_aws_version,
                                         edition_addon, ddc_dev_cfn_name)

    ddc_dev_no_vpc_cfn_name = "{}-ddc-dev-no-vpc".format(cfn_name)
    ddc_dev_url_no_vpc = create_cfn_template(DDCDevVPCExistingTemplate,
                                                ami_list, ddc_channel,
                                                docker_version, docker_for_aws_version,
                                                edition_addon, ddc_dev_no_vpc_cfn_name, cfn_type="no-vpc")

    if args.upload:
        print(u"Uploading templates.. \n")
        # CE upload
        s3_url = upload_cfn_template(release_channel, cfn_name, base_url)
        s3_url_no_vpc = upload_cfn_template(release_channel, no_vpc_cfn_name, base_url_no_vpc, cfn_type="no-vpc")
        # EE upload
        s3_docker_ee_url = upload_cfn_template(release_channel, docker_ee_cfn_name, docker_ee_base_url)
        s3_docker_ee_url_no_vpc = upload_cfn_template(release_channel, docker_ee_no_vpc_cfn_name, docker_ee_base_url_no_vpc, cfn_type="no-vpc")
        # Cloud upload
        s3_cloud_url = upload_cfn_template(release_channel, cloud_cfn_name, cloud_url)
        s3_cloud_url_no_vpc = upload_cfn_template(release_channel, cloud_no_vpc_cfn_name, cloud_url_no_vpc, cfn_type="no-vpc")
        # DDC upload
        s3_ddc_url = upload_cfn_template(release_channel, ddc_cfn_name, ddc_url)
        s3_ddc_url_no_vpc = upload_cfn_template(release_channel, ddc_no_vpc_cfn_name, ddc_url_no_vpc, cfn_type="no-vpc")
        # DDC dev upload
        s3_ddc_dev_url = upload_cfn_template(release_channel, ddc_dev_cfn_name, ddc_dev_url)
        s3_ddc_dev_url_no_vpc = upload_cfn_template(release_channel, ddc_dev_no_vpc_cfn_name, ddc_dev_url_no_vpc, cfn_type="no-vpc")
        print(u"Uploaded CFN URL={0} \n\t URL_NO_VPC={1}".format(s3_url, s3_url_no_vpc))
        print(u"Uploaded CFN EE_URL={0} \n\t EE_URL_NO_VPC={1}".format(s3_docker_ee_url, s3_docker_ee_url_no_vpc))
        print(u"Uploaded CFN CLOUD_URL={0} \n\t CLOUD_URL_NO_VPC={1}".format(s3_cloud_url, s3_cloud_url_no_vpc))
        print(u"Uploaded CFN DDC_URL={0} \n\t DDC_URL_NO_VPC={1}".format(s3_ddc_url, s3_ddc_url_no_vpc))
        print(u"Uploaded CFN DDC_DEV_URL={0} \n\t DDC_DEV_URL_NO_VPC={1}".format(s3_ddc_dev_url, s3_ddc_dev_url_no_vpc))

    # TODO: git commit, tag release. requires github keys, etc.

    print("------------------")
    print(u"Finshed.. Docker CE Base URL={}".format(base_url))
    print(u"Finshed.. Docker CE Base No VPC URL={}".format(base_url_no_vpc))
    print(u"Finshed.. Docker EE Base URL={}".format(docker_ee_base_url))
    print(u"Finshed.. Docker EE Base No VPC URL={}".format(docker_ee_base_url_no_vpc))
    print(u"Finshed.. Docker EE DDC URL={}".format(ddc_url))
    print(u"Finshed.. Docker EE DDC No VPC URL={}".format(ddc_url_no_vpc))
    print(u"Finshed.. Docker EE DDC Dev URL={}".format(ddc_dev_url))
    print(u"Finshed.. Docker EE DDC Dev No VPC URL={}".format(ddc_dev_url_no_vpc))
    print(u"Finshed.. CloudFormation Cloud URL={}".format(cloud_url))
    print(u"Finshed.. CloudFormation Cloud No VPC URL={}".format(cloud_url_no_vpc))
    print("Don't forget to tag the code (git tag -a {0} -m {0}; git push --tags)".format(docker_for_aws_version))
    print("------------------")


if __name__ == '__main__':
    main()
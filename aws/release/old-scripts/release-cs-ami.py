#!/usr/bin/env python
# given an AMI and account list, copy AMI to regions and share with accounts
import json
import argparse
import sys
from os import path
sys.path.append( path.dirname( path.dirname( path.abspath(__file__) ) ) )


from utils import (
    get_ami, wait_for_ami_to_be_available, copy_amis, strip_and_padd, get_account_list,
    approve_accounts, DOCKER_AWS_ACCOUNT_URL, upload_ami_list, create_cfn_template)


def main():
    parser = argparse.ArgumentParser(description='CS AMI Copy and approve')
    parser.add_argument('-d', '--docker_version',
                        dest='docker_version', required=True,
                        help="Docker version (i.e. 1.12.0-rc4)")
    parser.add_argument('-a', '--ami_id',
                        dest='ami_id', required=True,
                        help="ami-id for the Moby AMI we just built (i.e. ami-123456)")
    parser.add_argument('-s', '--ami_src_region',
                        dest='ami_src_region', required=True,
                        help="The reason the source AMI was built in (i.e. us-east-1)")
    parser.add_argument('-l', '--account_list_url',
                        dest='account_list_url', default=DOCKER_AWS_ACCOUNT_URL,
                        help="The URL for the aws account list for ami approvals")
    parser.add_argument('-p', '--public',
                        dest='make_ami_public', default="no",
                        help="Make the AMI public")

    args = parser.parse_args()

    docker_version = args.docker_version
    # TODO change to something else? where to get moby version?
    moby_version = docker_version
    docker_for_aws_ddc_version = u"aws-v{}-ddc".format(docker_version)
    image_name = u"Moby Linux {}".format(docker_for_aws_ddc_version)
    image_description = u"The best OS for running Docker, version {}".format(moby_version)
    print("\n Variables")
    print(u"docker_version={}".format(docker_version))
    print(u"ami_id={}".format(args.ami_id))
    print(u"ami_src_region={}".format(args.ami_src_region))
    print(u"account_list_url={}".format(args.account_list_url))
    print(u"make_ami_public={}".format(args.make_ami_public))
    if not args.account_list_url:
        print("account_list_url parameter is None, defaulting")
        account_list_url = DOCKER_AWS_ACCOUNT_URL
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
                         image_name, image_description, 'cs')
    ami_list_json = json.dumps(ami_list, indent=4, sort_keys=True)
    print(u"AMI copy complete. AMI List: \n{}".format(ami_list_json))

    print("upload CS AMI list to s3")
    upload_ami_list(ami_list_json, docker_version)

    print("Get account list..")
    account_list = get_account_list(account_list_url)
    print(u"Approving AMIs for {} accounts..".format(len(account_list)))
    approve_accounts(ami_list, account_list)
    print("Accounts have been approved.")

    print("------------------")
    print(u"Finshed copying AMI to all regions")
    print(ami_list_json)
    print("------------------")


if __name__ == '__main__':
    main()

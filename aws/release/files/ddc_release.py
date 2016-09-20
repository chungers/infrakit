#!/usr/bin/env python
import json
import argparse
import sys
from os import path
sys.path.append( path.dirname( path.dirname( path.abspath(__file__) ) ) )


from utils import (create_cfn_template, create_ddc_dev_cfn_template,
                   get_ami_list, CS_AMI_LIST_PATH, S3_BUCKET_NAME)

CFN_DDC_TEMPLATE = '/home/docker/docker_for_aws_ddc.template'


def main():
    parser = argparse.ArgumentParser(description='CS AMI Copy and approve')
    parser.add_argument('-d', '--docker_version',
                        dest='docker_version', required=True,
                        help="Docker version (i.e. 1.12.0-rc4)")
    parser.add_argument('-e', '--edition_version',
                        dest='edition_version', required=True,
                        help="Edition version (i.e. Beta 4)")
    parser.add_argument('-c', '--channel',
                        dest='channel', default="beta",
                        help="release channel (beta, alpha, rc, nightly)")

    args = parser.parse_args()

    docker_version = args.docker_version
    edition_version = args.edition_version
    channel = args.channel

    # TODO change to something else? where to get moby version?
    moby_version = docker_version
    flat_edition_version = edition_version.replace(" ", "")
    docker_for_aws_ddc_version = u"aws-v{}-{}-ddc".format(docker_version, flat_edition_version)
    docker_for_aws_ddc_dev_version = u"{}-dev".format(docker_for_aws_ddc_version)
    image_name = u"Moby Linux {}".format(docker_for_aws_ddc_version)
    image_description = u"The best OS for running Docker, version {}".format(moby_version)
    print("\n Variables")
    print(u"docker_version={}".format(docker_version))
    print(u"edition_version={}".format(edition_version))
    print(u"channel={}".format(channel))

    ami_list_path = CS_AMI_LIST_PATH.format(docker_version)
    ami_list_url = u"https://{}.s3.amazonaws.com/{}".format(S3_BUCKET_NAME, ami_list_path)
    ami_list_json = get_ami_list(ami_list_url)

    print("DDC release")
    print(u"CS AMI List: \n{}".format(ami_list_json))
    print("Create DDC CloudFormation template..")
    s3_ddc_url = create_cfn_template(ami_list_json, channel, docker_version,
                                     docker_for_aws_ddc_version, edition_version, CFN_DDC_TEMPLATE,
                                     docker_for_aws_ddc_version)

    print("------------------")
    print(u"CloudFormation DDC_URL={} \n".format(s3_ddc_url))
    print("------------------")

    print("DDC DEV release")
    print(u"CS AMI List: \n{}".format(ami_list_json))
    print("Create DDC CloudFormation template..")
    s3_ddc_dev_url = create_ddc_dev_cfn_template(ami_list_json, channel, docker_version,
                                     docker_for_aws_ddc_version, edition_version, CFN_DDC_TEMPLATE,
                                     docker_for_aws_ddc_dev_version)

    print("------------------")
    print(u"Finshed.. CloudFormation DEV DDC_URL={} \n".format(s3_ddc_dev_url))
    print("------------------")


    print("------------------")
    print(u"Finshed copying AMI to all regions")
    print("------------------")


if __name__ == '__main__':
    main()

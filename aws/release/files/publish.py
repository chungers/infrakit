#!/usr/bin/env python
import json
import argparse
import re

from utils import publish_cfn_template


CFN_TEMPLATE = '/home/docker/docker_for_aws.template'
CFN_CLOUD_TEMPLATE = '/home/docker/docker_for_aws_cloud.template'


def main():
    parser = argparse.ArgumentParser(description='Release Docker for AWS')
    parser.add_argument('-r', '--release_version',
                        dest='release_version', required=True,
                        help="Docker release version (i.e. 1.12.3-beta10)")
    parser.add_argument('-c', '--channel',
                        dest='channel', default="beta", required=True,
                        help="release channel (beta, alpha, rc, nightly)")
    args = parser.parse_args()

    release_version = args.release_version
    release_channel = args.channel
    docker_for_aws_version = u"aws-v{}".format(release_version)
    print("\nVariables")
    print(u"release_channel={}".format(release_channel))
    print(u"release_version={}".format(release_version))

    s3_latest_url = publish_cfn_template(release_channel, docker_for_aws_version)
    print("------------------")
    print(u"Finshed.. Latest CloudFormation \n\t URL={0}".format(s3_latest_url))
    print("------------------")

if __name__ == '__main__':
    main()

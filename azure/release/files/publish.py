#!/usr/bin/env python
import json
import argparse
import re

from utils import publish_rg_template


CFN_TEMPLATE = '/home/docker/editions.template'


def main():
    parser = argparse.ArgumentParser(description='Release Docker for AWS')
    parser.add_argument('-r', '--release_version',
                        dest='release_version', required=True,
                        help="Docker release version (i.e. 1.12.3-beta10)")
    parser.add_argument('-c', '--channel',
                        dest='channel', default="beta", required=True,
                        help="release channel (beta, alpha, rc, nightly)")
    args = parser.parse_args()

    docker_for_azure_version = args.release_version
    release_channel = args.channel
    print("\nVariables")
    print(u"release_channel={}".format(release_channel))
    print(u"docker_for_azure_version={}".format(docker_for_azure_version))

    s3_latest_url = publish_rg_template(release_channel, docker_for_azure_version)
    print("------------------")
    print(u"Finshed.. Latest Azure template \n\t URL={0}".format(s3_latest_url))
    print("------------------")

if __name__ == '__main__':
    main()

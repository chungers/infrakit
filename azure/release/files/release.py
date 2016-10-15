#!/usr/bin/env python
import json
import argparse

from utils import (create_cfn_template)


CFN_TEMPLATE = '/home/docker/editions.template'

def main():
    parser = argparse.ArgumentParser(description='Release Docker for AWS')
    parser.add_argument('-d', '--docker_version',
                        dest='docker_version', required=True,
                        help="Docker version (i.e. 1.12.0-rc4)")
    parser.add_argument('-e', '--edition_version',
                        dest='edition_version', required=True,
                        help="Edition version (i.e. Beta 4)")
    parser.add_argument('-s', '--vhd_sku',
                        dest='vhd_sku', required=True,
                        help="The Azure VHD SKU (i.e. docker4azure)")
    parser.add_argument('-v', '--vhd_version',
                        dest='vhd_version', required=True,
                        help="The Azure VHD version (i.e. 1.0.0)")
    parser.add_argument('-c', '--channel',
                        dest='channel', default="beta",
                        help="release channel (beta, alpha, rc, nightly)")

    args = parser.parse_args()

    release_channel = args.channel
    docker_version = args.docker_version
    # TODO change to something else? where to get moby version?
    moby_version = docker_version
    edition_version = args.edition_version
    flat_edition_version = edition_version.replace(" ", "")
    vhd_sku = args.vhd_sku
    vhd_version = args.vhd_version
    docker_for_azure_version = u"azure-v{}-{}".format(docker_version, flat_edition_version)
    image_name = u"Moby Linux {}".format(docker_for_azure_version)
    image_description = u"The best OS for running Docker, version {}".format(moby_version)
    print("\n Variables")
    print(u"release_channel={}".format(release_channel))
    print(u"docker_version={}".format(docker_version))
    print(u"edition_version={}".format(edition_version))
    print(u"vhd_sku={}".format(vhd_sku))
    print(u"vhd_version={}".format(vhd_version))

    print("Create CloudFormation template..")
    s3_url = create_cfn_template(vhd_sku, vhd_version, release_channel, docker_version,
                                 docker_for_azure_version, edition_version, CFN_TEMPLATE,
                                 docker_for_azure_version)

    # TODO: git commit, tag release. requires github keys, etc.

    print("------------------")
    print(u"Finshed.. CloudFormation URL={0} \n".format(s3_url))
    print("Don't forget to tag the code (git tag -a v{0}-{1} -m {0}; git push --tags)".format(
        docker_version, flat_edition_version))
    print("------------------")


if __name__ == '__main__':
    main()

#!/usr/bin/env python
import json
import argparse

from utils import (create_rg_template, create_rg_cloud_template, create_rg_ddc_template, upload_rg_template)


CFN_TEMPLATE = '/home/docker/editions.template'

TEMPLATE_EXTENSION = u".tmpl"
AZURE_PLATFORMS = {
    "PUBLIC" : {
        "STORAGE_ENDPOINT" : u".blob.core.windows.net",
        "PORTAL_ENDPOINT"  : u"portal.azure.com",
        "TEMPLATE_SUFFIX"  : ""
    },
    "GOV" : {
        "STORAGE_ENDPOINT" : u".blob.core.usgovcloudapi.net",
        "PORTAL_ENDPOINT"  : u"portal.azure.us",
        "TEMPLATE_SUFFIX"  : u"-gov"
    }
}

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
                        help="The Azure VHD SKU (i.e. docker-ce)")
    parser.add_argument('-v', '--vhd_version',
                        dest='vhd_version', required=True,
                        help="The Azure VHD version (i.e. 1.0.0)")
    parser.add_argument('-c', '--channel',
                        dest='channel', default="beta",
                        help="release channel (beta, alpha, rc, nightly)")
    parser.add_argument('-u', '--channel_cloud',
                        dest='channel_cloud', default="alpha",
                        help="Cloud release channel (beta, alpha, rc, nightly)")
    parser.add_argument('--channel_ddc',
                        dest='channel_ddc', default="alpha",
                        help="DDC release channel (beta, alpha, rc, nightly)")
    parser.add_argument('--offer_id',
                        dest='offer_id', default="docker-ce",
                        help="The Azure VHD Offer ID")
    parser.add_argument('--cs_vhd_sku',
                        dest='cs_vhd_sku',
                        help="The Azure CS VHD SKU (i.e. docker-ce)")
    parser.add_argument('--cs_vhd_version',
                        dest='cs_vhd_version',
                        help="The Azure CS VHD version (i.e. 1.0.0)")
    parser.add_argument('--cs_offer_id',
                        dest='cs_offer_id', default="docker-ee",
                        help="The Azure CS VHD Offer ID")
    parser.add_argument("--upload", action="store_true",
                        help="Upload the Azure template once generated")

    args = parser.parse_args()

    release_channel = args.channel
    release_cloud_channel = args.channel_cloud
    release_ddc_channel = args.channel_cloud
    docker_version = args.docker_version
    # TODO change to something else? where to get moby version?
    moby_version = docker_version
    edition_version = args.edition_version
    flat_edition_version = edition_version.replace(" ", "")
    vhd_sku = args.vhd_sku
    vhd_version = args.vhd_version
    offer_id = args.offer_id
    cs_vhd_sku = args.cs_vhd_sku
    cs_vhd_version = args.cs_vhd_version
    cs_offer_id = args.cs_offer_id

    docker_for_azure_version = u"azure-v{}".format(flat_edition_version)
    image_name = u"Moby Linux {}".format(docker_for_azure_version)
    image_description = u"The best OS for running Docker, version {}".format(moby_version)
    print("\n Variables")
    print(u"release_channel={}".format(release_channel))
    print(u"release_cloud_channel={}".format(release_cloud_channel))
    print(u"release_ddc_channel={}".format(release_ddc_channel))
    print(u"docker_version={}".format(docker_version))
    print(u"edition_version={}".format(edition_version))
    print(u"vhd_sku={}".format(vhd_sku))
    print(u"vhd_version={}".format(vhd_version))
    print(u"cs_vhd_sku={}".format(cs_vhd_sku))
    print(u"cs_vhd_version={}".format(cs_vhd_version))

    print("Create ARM templates..")
    for platform, platform_config in AZURE_PLATFORMS.items():
        template_name = u"Docker" + platform_config['TEMPLATE_SUFFIX'] + TEMPLATE_EXTENSION
        base_url = create_rg_template(vhd_sku, vhd_version, offer_id, release_channel, docker_version,
                                 docker_for_azure_version, edition_version, CFN_TEMPLATE,
                                 platform_config['STORAGE_ENDPOINT'],
                                 platform_config['PORTAL_ENDPOINT'],
                                 template_name)
        cloud_template_name = u"Docker-Cloud" + platform_config['TEMPLATE_SUFFIX'] + TEMPLATE_EXTENSION
        cloud_url = create_rg_cloud_template(release_cloud_channel, docker_version,
                                 docker_for_azure_version, edition_version, base_url,
                                 platform_config['STORAGE_ENDPOINT'],
                                 platform_config['PORTAL_ENDPOINT'],
                                 cloud_template_name)

        ddc_template_name = u"Docker-DDC" + platform_config['TEMPLATE_SUFFIX'] + TEMPLATE_EXTENSION
        ddc_url = create_rg_ddc_template(cs_vhd_sku, cs_vhd_version, cs_offer_id, release_ddc_channel, docker_version,
                                 docker_for_azure_version, edition_version, base_url,
                                 platform_config['STORAGE_ENDPOINT'],
                                 platform_config['PORTAL_ENDPOINT'],
                                 ddc_template_name)

        print("------------------")

        if args.upload:
            print(u"Uploading templates.. \n")
            s3_url = upload_rg_template(release_channel, template_name, base_url)
            s3_cloud_url = upload_rg_template(release_channel, cloud_template_name, cloud_url)
            s3_ddc_url = upload_rg_template(release_channel, ddc_template_name, ddc_url)
            print(u"Uploaded ARM \n\t URL={0} \n\t CLOUD_URL={1} \n\t DDC_URL={2} \n".format(s3_url, s3_cloud_url, s3_ddc_url))

    print(u"Finshed.. \n")

    # TODO: git commit, tag release. requires github keys, etc.
    print("Don't forget to tag the code (git tag -a v{0} -m {1}; git push --tags)".format(
        edition_version, docker_for_azure_version))
    print("------------------")


if __name__ == '__main__':
    main()

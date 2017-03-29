#!/usr/bin/env python

# Only used for testing, to generate templates without creating the AMIs

import json
import os

from base import AWSBaseTemplate
from existing_vpc import ExistingVPCTemplate
from cloud import CloudVPCTemplate, CloudVPCExistingTemplate
from docker_ee import DockerEEVPCTemplate, DockerEEVPCExistingTemplate
from ddc import DDCVPCTemplate, DDCVPCExistingTemplate
from ddc_dev import DDCDevVPCTemplate, DDCDevVPCExistingTemplate


def generate_template(template_class, docker_version,
                      docker_for_aws_version,
                      edition_addon, channel, amis, file_name):
    aws_template = template_class(docker_version, docker_for_aws_version,
                                  edition_addon, channel, amis)
    aws_template.build()

    # TODO: redudant open json to just write to file.
    new_template = json.loads(aws_template.generate_template())

    curr_path = os.path.dirname(__file__)
    out_path = os.path.join(curr_path, 'outputs/{}'.format(file_name))
    with open(out_path, 'w') as newfile:
        json.dump(new_template, newfile, sort_keys=True,
                  indent=4, separators=(',', ': '))


if __name__ == '__main__':
    docker_version = "17.03.1-ce-rc1"
    channel = "stable"
    docker_for_aws_version = '{}-aws1'.format(docker_version)

    amis = {
            "ap-northeast-1": {
                "HVM64": "ami-35237952",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ap-northeast-2": {
                "HVM64": "ami-795b8817",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ap-south-1": {
                "HVM64": "ami-cb0675a4",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ap-southeast-1": {
                "HVM64": "ami-c0932fa3",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ap-southeast-2": {
                "HVM64": "ami-5856593b",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ca-central-1": {
                "HVM64": "ami-ce18a5aa",
                "HVMG2": "NOT_SUPPORTED"
            },
            "eu-central-1": {
                "HVM64": "ami-672eff08",
                "HVMG2": "NOT_SUPPORTED"
            },
            "eu-west-1": {
                "HVM64": "ami-3e98a558",
                "HVMG2": "NOT_SUPPORTED"
            },
            "eu-west-2": {
                "HVM64": "ami-bf7c68db",
                "HVMG2": "NOT_SUPPORTED"
            },
            "sa-east-1": {
                "HVM64": "ami-8dc1a1e1",
                "HVMG2": "NOT_SUPPORTED"
            },
            "us-east-1": {
                "HVM64": "ami-b079c7a6",
                "HVMG2": "NOT_SUPPORTED"
            },
            "us-east-2": {
                "HVM64": "ami-a91d39cc",
                "HVMG2": "NOT_SUPPORTED"
            },
            "us-west-1": {
                "HVM64": "ami-a1bae1c1",
                "HVMG2": "NOT_SUPPORTED"
            },
            "us-west-2": {
                "HVM64": "ami-d849dcb8",
                "HVMG2": "NOT_SUPPORTED"
            }
    }

    # Docker CE
    generate_template(ExistingVPCTemplate, docker_version, docker_for_aws_version,
                      'base' ,channel, amis,
                      'docker_ce_for_aws_no_vpc.json')
    generate_template(AWSBaseTemplate, docker_version, docker_for_aws_version,
                      'base', channel, amis,
                      'docker_ce_for_aws.json')

    # Docker EE
    generate_template(DockerEEVPCExistingTemplate, docker_version,
                      docker_for_aws_version, 'base-ee', channel, amis,
                      'docker_ee_for_aws_no_vpc.json')
    generate_template(DockerEEVPCTemplate, docker_version, docker_for_aws_version,
                      'base-ee', channel, amis,
                      'docker_ee_for_aws.json')

    # Docker Cloud
    generate_template(CloudVPCExistingTemplate, docker_version, docker_for_aws_version,
                      'cloud', channel, amis,
                      'docker_ce_for_aws_no_vpc_cloud.json')
    generate_template(CloudVPCTemplate, docker_version, docker_for_aws_version,
                      'cloud', channel, amis,
                      'docker_ce_for_aws_cloud.json')

    # Docker DDC
    generate_template(DDCVPCExistingTemplate, docker_version, docker_for_aws_version,
                      'ddc', channel, amis,
                      'docker_ee_for_aws_no_vpc_ddc.json')
    generate_template(DDCVPCTemplate, docker_version, docker_for_aws_version,
                      'ddc', channel, amis,
                      'docker_ee_for_aws_ddc.json')

    # Docker DDC Dev
    generate_template(DDCDevVPCExistingTemplate, docker_version,
                      docker_for_aws_version, 'ddc', channel, amis,
                      'docker_ee_for_aws_no_vpc_ddc_dev.json')
    generate_template(DDCDevVPCTemplate, docker_version,
                      docker_for_aws_version, 'ddc', channel, amis,
                      'docker_ee_for_aws_ddc_dev.json')

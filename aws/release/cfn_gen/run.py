#!/usr/bin/env python

# Only used for testing, to generate templates without creating the AMIs

import json
import os

from docker_ce import DockerCEVPCTemplate, DockerCEVPCExistingTemplate
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
    docker_version = "17.05.0-ce"
    channel = "edge"
    docker_for_aws_version = '{}-aws1'.format(docker_version)

    amis = {
        "ap-northeast-1": {
            "HVM64": "ami-e79dac80",
            "HVMG2": "NOT_SUPPORTED"
        },
        "ap-northeast-2": {
            "HVM64": "ami-83d508ed",
            "HVMG2": "NOT_SUPPORTED"
        },
        "ap-south-1": {
            "HVM64": "ami-eacab885",
            "HVMG2": "NOT_SUPPORTED"
        },
        "ap-southeast-1": {
            "HVM64": "ami-37de5b54",
            "HVMG2": "NOT_SUPPORTED"
        },
        "ap-southeast-2": {
            "HVM64": "ami-93fcf7f0",
            "HVMG2": "NOT_SUPPORTED"
        },
        "ca-central-1": {
            "HVM64": "ami-658c3001",
            "HVMG2": "NOT_SUPPORTED"
        },
        "eu-central-1": {
            "HVM64": "ami-f66fb099",
            "HVMG2": "NOT_SUPPORTED"
        },
        "eu-west-1": {
            "HVM64": "ami-26030f40",
            "HVMG2": "NOT_SUPPORTED"
        },
        "eu-west-2": {
            "HVM64": "ami-5fa3b73b",
            "HVMG2": "NOT_SUPPORTED"
        },
        "sa-east-1": {
            "HVM64": "ami-319af65d",
            "HVMG2": "NOT_SUPPORTED"
        },
        "us-east-1": {
            "HVM64": "ami-17b2db01",
            "HVMG2": "NOT_SUPPORTED"
        },
        "us-east-2": {
            "HVM64": "ami-1be5c27e",
            "HVMG2": "NOT_SUPPORTED"
        },
        "us-west-1": {
            "HVM64": "ami-8b4264eb",
            "HVMG2": "NOT_SUPPORTED"
        },
        "us-west-2": {
            "HVM64": "ami-6166fd01",
            "HVMG2": "NOT_SUPPORTED"
        }
    }

    # Docker CE
    generate_template(DockerCEVPCExistingTemplate, docker_version,
                      docker_for_aws_version,
                      'base', channel, amis,
                      'docker_ce_for_aws_no_vpc.json')
    generate_template(DockerCEVPCTemplate, docker_version,
                      docker_for_aws_version,
                      'base', channel, amis,
                      'docker_ce_for_aws.json')

    # Docker EE
    generate_template(DockerEEVPCExistingTemplate, docker_version,
                      docker_for_aws_version, 'base-ee', channel, amis,
                      'docker_ee_for_aws_no_vpc.json')
    generate_template(DockerEEVPCTemplate, docker_version,
                      docker_for_aws_version,
                      'base-ee', channel, amis,
                      'docker_ee_for_aws.json')

    # Docker Cloud
    generate_template(CloudVPCExistingTemplate, docker_version,
                      docker_for_aws_version,
                      'cloud', channel, amis,
                      'docker_ce_for_aws_no_vpc_cloud.json')
    generate_template(CloudVPCTemplate, docker_version, docker_for_aws_version,
                      'cloud', channel, amis,
                      'docker_ce_for_aws_cloud.json')

    # Docker DDC
    generate_template(DDCVPCExistingTemplate, docker_version,
                      docker_for_aws_version,
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

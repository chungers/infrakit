#!/usr/bin/env python

# Only used for testing, to generate templates without creating the AMIs

import json
import os

from base import AWSBaseTemplate
from existing_vpc import ExistingVPCTemplate
from cloud import CloudVPCTemplate, CloudVPCExistingTemplate
from docker_ee import DockerEEVPCTemplate, DockerEEVPCExistingTemplate
from ddc import DDCVPCTemplate, DDCVPCExistingTemplate


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

    AMI = "ami-57cc1341"
    amis = {"ap-northeast-1": {
                "HVM64": "ami-b37e2bd4",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ap-northeast-2": {
                "HVM64": "ami-923dedfc",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ap-south-1": {
                "HVM64": "ami-d21b6bbd",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ap-southeast-1": {
                "HVM64": "ami-ab1baac8",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ap-southeast-2": {
                "HVM64": "ami-253c3f46",
                "HVMG2": "NOT_SUPPORTED"
            },
            "ca-central-1": {
                "HVM64": "ami-ad72cfc9",
                "HVMG2": "NOT_SUPPORTED"
            },
            "eu-central-1": {
                "HVM64": "ami-2acd1845",
                "HVMG2": "NOT_SUPPORTED"
            },
            "eu-west-1": {
                "HVM64": "ami-a5ad82c3",
                "HVMG2": "NOT_SUPPORTED"
            },
            "eu-west-2": {
                "HVM64": "ami-7cc5d018",
                "HVMG2": "NOT_SUPPORTED"
            },
            "sa-east-1": {
                "HVM64": "ami-3babcd57",
                "HVMG2": "NOT_SUPPORTED"
            },
            "us-east-1": {
                "HVM64": "ami-57cc1341",
                "HVMG2": "NOT_SUPPORTED"
            },
            "us-east-2": {
                "HVM64": "ami-feb89d9b",
                "HVMG2": "NOT_SUPPORTED"
            },
            "us-west-1": {
                "HVM64": "ami-0e67396e",
                "HVMG2": "NOT_SUPPORTED"
            },
            "us-west-2": {
                "HVM64": "ami-21c94b41",
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

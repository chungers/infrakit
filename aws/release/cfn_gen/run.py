#!/usr/bin/env python

# Only used for testing, to generate templates without creating the AMIs

import json
import os

from base import AWSBaseTemplate
from existing_vpc import ExistingVPCTemplate
from cloud import CloudVPCTemplate, CloudVPCExistingTemplate
from docker_ee import DockerEEVPCTemplate, DockerEEVPCExistingTemplate


def generate_template(template_class, docker_version,
                      edition_version, docker_for_aws_version,
                      channel, amis, file_name):
    aws_template = template_class(docker_version, edition_version,
                                  docker_for_aws_version,
                                  channel, amis)
    aws_template.build()

    # TODO: redudant open json to just write to file.
    new_template = json.loads(aws_template.generate_template())

    curr_path = os.path.dirname(__file__)
    out_path = os.path.join(curr_path, 'outputs/{}'.format(file_name))
    with open(out_path, 'w') as newfile:
        json.dump(new_template, newfile, sort_keys=True,
                  indent=4, separators=(',', ': '))


if __name__ == '__main__':
    docker_version = "17.03.0-ce-rc1"
    edition_version = "beta19"
    channel = "beta"
    docker_for_aws_version = 'aws-v17.03.0-ce-rc1-beta19'

    AMI = "ami-2b93403d"
    amis = {"us-east-1": {
            "HVM64": AMI,
            "HVMG2": "NOT_SUPPORTED"}}

    # Docker CE
    generate_template(ExistingVPCTemplate, docker_version, edition_version,
                      docker_for_aws_version, channel, amis,
                      'docker_ce_for_aws_no_vpc.json')
    generate_template(AWSBaseTemplate, docker_version, edition_version,
                      docker_for_aws_version, channel, amis,
                      'docker_ce_for_aws.json')

    # Docker EE
    generate_template(DockerEEVPCExistingTemplate, docker_version,
                      edition_version, docker_for_aws_version, channel, amis,
                      'docker_ee_for_aws_no_vpc.json')
    generate_template(DockerEEVPCTemplate, docker_version, edition_version,
                      docker_for_aws_version, channel, amis,
                      'docker_ee_for_aws.json')

    # Docker Cloud
    generate_template(CloudVPCExistingTemplate, docker_version, edition_version,
                      docker_for_aws_version, channel, amis,
                      'docker_ce_for_aws_no_vpc_cloud.json')
    generate_template(CloudVPCTemplate, docker_version, edition_version,
                      docker_for_aws_version, channel, amis,
                      'docker_ce_for_aws_cloud.json')

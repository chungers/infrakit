#!/usr/bin/env python

# Only used for testing, will be removed once testing is complete.

import json
import os

from base import AWSBaseTemplate
from existing_vpc import ExistingVPCTemplate
from cloud import CloudVPCTemplate, CloudVPCExistingTemplate


def ordered(obj):
    if isinstance(obj, dict):
        return sorted((k, ordered(v)) for k, v in obj.items())
    if isinstance(obj, list):
        return sorted(ordered(x) for x in obj)
    else:
        return obj


def generate_base_template(docker_version, edition_version, channel, amis):
    aws_template = AWSBaseTemplate(docker_version, edition_version, channel, amis)
    aws_template.build()

    # TODO: redudant open json to just write to file.
    new_template = json.loads(aws_template.generate_template())
    curr_path = os.path.dirname(__file__)
    out_path = os.path.join(curr_path, 'outputs/docker_for_aws.json')

    with open(out_path, 'w') as newfile:
        json.dump(new_template, newfile, sort_keys=True, indent=4, separators=(',', ': '))


def generate_existing_VPC_template(docker_version, edition_version, channel, amis):
    aws_template = ExistingVPCTemplate(docker_version, edition_version, channel, amis)
    aws_template.build()

    # TODO: redudant open json to just write to file.
    new_template = json.loads(aws_template.generate_template())

    curr_path = os.path.dirname(__file__)
    out_path = os.path.join(curr_path, 'outputs/docker_for_aws_no_vpc.json')
    with open(out_path, 'w') as newfile:
        json.dump(new_template, newfile, sort_keys=True, indent=4, separators=(',', ': '))


def generate_cloud_VPC_template(docker_version, edition_version, channel, amis):
    aws_template = CloudVPCTemplate(docker_version, edition_version, channel, amis)
    aws_template.build()

    # TODO: redudant open json to just write to file.
    new_template = json.loads(aws_template.generate_template())

    curr_path = os.path.dirname(__file__)
    out_path = os.path.join(curr_path, 'outputs/docker_for_aws_cloud.json')
    with open(out_path, 'w') as newfile:
        json.dump(new_template, newfile, sort_keys=True, indent=4, separators=(',', ': '))


def generate_cloud_NO_VPC_template(docker_version, edition_version, channel, amis):
    aws_template = CloudVPCExistingTemplate(docker_version, edition_version, channel, amis)
    aws_template.build()

    # TODO: redudant open json to just write to file.
    new_template = json.loads(aws_template.generate_template())

    curr_path = os.path.dirname(__file__)
    out_path = os.path.join(curr_path, 'outputs/docker_for_aws_no_vpc_cloud.json')
    with open(out_path, 'w') as newfile:
        json.dump(new_template, newfile, sort_keys=True, indent=4, separators=(',', ': '))


if __name__ == '__main__':
    docker_version = "1.13.1-rc1"
    edition_version = "beta 16"
    channel = "beta"
    curr_path = os.path.dirname(__file__)
    ami_path = os.path.join(curr_path, 'amis.json')
    print(ami_path)
    try:
        with open(ami_path) as data_file:
            amis = json.load(data_file)
    except Exception as e:
        print(e)
        raise e

    generate_existing_VPC_template(docker_version, edition_version, channel, amis)
    generate_base_template(docker_version, edition_version, channel, amis)
    generate_cloud_VPC_template(docker_version, edition_version, channel, amis)
    generate_cloud_NO_VPC_template(docker_version, edition_version, channel, amis)

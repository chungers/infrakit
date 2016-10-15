#!/usr/bin/env python
import os
import json
import time
from datetime import datetime
import boto
import urllib2
from boto import ec2
import re
from collections import OrderedDict

NOW = datetime.now()
NOW_STRING = NOW.strftime("%m_%d_%Y")

AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
S3_BUCKET_NAME = "docker-for-azure"

def str2bool(v):
  return v.lower() in ("yes", "true", "t", "1")

def is_json(cfn_template):
  try:
    with open(cfn_template) as data_file:
      json_object = json.load(data_file)
  except ValueError, e:
    return False
  return True

def buildCustomData(data_file):
  customData = []
  with open(data_file) as f:
    for line in f:
      m = re.match(r'(.*?)((?:parameters|variables)\([^\)]*\))(.*$)', line)
      if m:
        customData += ['\'' + m.group(1) + '\'',
              m.group(2),
              '\'' + m.group(3) + '\'',
              '\'\n\'']
      else:
        customData += ['\'' + line + '\'']
  return customData

def upload_cfn_template(release_channel, cloudformation_template_name, tempfile, cfn_type=''):

    # upload to s3, make public, return s3 URL
    s3_host_name = u"https://{}.s3.amazonaws.com".format(S3_BUCKET_NAME)
    s3_path = u"azure/{}/{}".format(release_channel, cloudformation_template_name)
    latest_name = "latest.json"
    if cfn_type:
        latest_name = "{}-latest.json".format(cfn_type)

    s3_path_latest = u"aws/{}/{}".format(release_channel, latest_name)
    s3_full_url = u"{}/{}".format(s3_host_name, s3_path)

    s3conn = boto.connect_s3(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(S3_BUCKET_NAME)

    print(u"Upload Cloudformation template to {} in {} s3 bucket".format(s3_path, S3_BUCKET_NAME))
    key = bucket.new_key(s3_path)
    key.set_metadata("Content-Type", "application/json")
    key.set_contents_from_filename(tempfile)
    key.set_acl("public-read")

    if release_channel == 'nightly' or release_channel == 'ddc-nightly'  or release_channel == 'cloud-nightly':
        print("This is a nightly build, update the latest.json file.")
        print(u"Upload Cloudformation template to {} in {} s3 bucket".format(
            s3_path_latest, S3_BUCKET_NAME))
        key = bucket.new_key(s3_path_latest)
        key.set_metadata("Content-Type", "application/json")
        key.set_contents_from_filename(tempfile)
        key.set_acl("public-read")

    return s3_full_url

def create_cfn_template(vhd_sku, vhd_version, release_channel, docker_version,
                        docker_for_azure_version, edition_version, cfn_template, cfn_name):
    # check if file exists before opening.
    flat_edition_version = edition_version.replace(" ", "").replace("_", "").replace("-", "")
    flat_edition_version_upper = flat_edition_version.capitalize()

    with open(cfn_template) as data_file:
        data = json.load(data_file)
    
    data['description'] = u"Docker for Azure {0} ({1})".format(docker_version, edition_version)
    data['variables']['imageSku'] = vhd_sku
    data['variables']['imageVersion'] = vhd_version
    data['variables']['dockerForIAASVersion'] = docker_for_azure_version

    # Updated Manager custom data
    manager_data = buildCustomData('custom-data_manager.sh')
    data['variables']['customDataManager'] = '[concat(' + ', '.join(manager_data) + ')]'
    # Updated Manager custom data
    worker_data = buildCustomData('custom-data_worker.sh')
    data['variables']['customDataWorker'] = '[concat(' + ', '.join(worker_data) + ')]'


    cloudformation_template_name = u"{}.json".format(cfn_name)
    outdir = u"dist/azure/{}".format(release_channel)
    # if the directory doesn't exist, create it.
    if not os.path.exists(outdir):
        os.makedirs(outdir)

    outfile = u"{}/{}".format(outdir, cloudformation_template_name)

    with open(outfile, 'w') as outf:
        json.dump(data, outf, indent=4, sort_keys=True)

    # TODO: json validate
    if is_json(outfile):
      print(u"Cloudformation template created in {}".format(outfile))
    else:
      print(u"ERROR: Cloudformation template invalid in {}".format(outfile))

    return upload_cfn_template(release_channel, cloudformation_template_name, outfile)



def create_ddc_dev_cfn_template(amis, release_channel, docker_version,
                                docker_for_aws_version, edition_version, cfn_template, cfn_name):
    """This will create the ddc dev template, which is for internal use only."""
    # check if file exists before opening.

    flat_edition_version = edition_version.replace(" ", "").replace("_", "").replace("-", "")
    flat_edition_version_upper = flat_edition_version.capitalize()

    with open(cfn_template) as data_file:
        data = json.load(data_file)

    parameters = data.get('Parameters')
    if parameters:
        new_parameters = {
            "HubNamespaceSet": {
                "Type": "String",
                "Description": "Hub namespace to pull Datacenter?",
                "AllowedValues": ["docker", "dockerorcadev"],
                "Default": "dockerorcadev"
            },
            "HubTagSet": {
                "Type": "String",
                "Description": "Hub tag to pull Datacenter?",
                "ConstraintDescription": "Please enter the image tag you want to use for pulling Docker Datacenter",
                "Default": "2.0.0-beta1"
            },
            "DockerIdSet": {
                "Type": "String",
                "Description": "Docker ID for pulling Datacenter image?",
                "ConstraintDescription": "A Docker account that has access to the specified Datacenter image on Docker Hub"
            },
            "DockerIdPasswordSet": {
                "Type": "String",
                "Description": "Docker password pulling image?",
                "ConstraintDescription": "Docker password corresponding to the Docker ID",
                "NoEcho": "true"
            }
        }
        parameters.update(new_parameters)

    outputs = data.get('Outputs')
    if outputs:
        new_outputs = {
            "TheHubNamespace": {
                "Description": "Docker Hub namespace",
                "Value": {
                    "Ref": "HubNamespaceSet"
                }
            },
            "TheHubTag": {
                "Description": "Docker Hub tag",
                "Value": {
                    "Ref": "HubTagSet"
                }
            },
            "TheDockerId": {
                "Description": "Docker ID",
                "Value": {
                    "Ref": "DockerIdSet"
                }
            },
        }
        outputs.update(new_outputs)

    metadata = data.get('Metadata')
    if metadata:
        new_optional_parameters = ["HubNamespaceSet", "HubTagSet", "DockerIdSet", "DockerIdPasswordSet"]
        new_parameter_labels = {
            "DockerIdSet": {"default": "Docker ID for installing Docker Datacenter?"},
            "DockerIdPasswordSet": {"default": "docker ID Password for installing Docker Datacenter?"},
            "HubNamespaceSet": {"default": "Hub namespace for installing Docker Datacenter?"},
            "HubTagSet": {"default": "Hub tag for installing Docker Datacenter?"},
        }
        try:
            for group in metadata["AWS::CloudFormation::Interface"]["ParameterGroups"]:
                if group["Label"]["default"] == "Optional Features":
                    group["Parameters"] += new_optional_parameters
            metadata["AWS::CloudFormation::Interface"]['ParameterLabels'].update(new_parameter_labels)
        except ValueError as e:
            print("Updating metadata failed: {}", e.args)

    data['Description'] = u"Docker for AWS {0} ({1}) DEV".format(docker_version, edition_version)
    data['Mappings']['DockerForAWS'] = {u'version': {u'docker': docker_version,
                                                     u'forAws': docker_for_aws_version}}
    data['Mappings']['AWSRegionArch2AMI'] = amis

    manager_launch_config_orig_key = 'ManagerLaunchConfig'
    manager_launch_config_new_key = u"{}{}".format(
        manager_launch_config_orig_key, flat_edition_version_upper)
    node_launch_config_orig_key = 'NodeLaunchConfig'
    node_launch_config_new_key = u"{}{}".format(
        node_launch_config_orig_key, flat_edition_version_upper)

    # loop through the resouces finding the correct Config node names.
    for key in data.get('Resources'):
        if key.startswith(manager_launch_config_orig_key):
            # we found the full name, change variable to current key value.
            manager_launch_config_orig_key = key
        elif key.startswith(node_launch_config_orig_key):
            # we found the full name, change variable to current key value.
            node_launch_config_orig_key = key

    docker_launch_command = 'docker4x/ddc-init-aws:$DOCKER_FOR_IAAS_VERSION\n'
    new_content = ["-e HUB_NAMESPACE='", {"Ref": "HubNamespaceSet"}, "' ",
                   "-e HUB_TAG='", {"Ref": "HubTagSet"}, "' ",
                   "-e DOCKER_ID='", {"Ref": "DockerIdSet"}, "' ",
                   "-e DOCKER_ID_PASSWORD='", {"Ref": "DockerIdPasswordSet"}, "' ",
                  ]

    data['Resources'][manager_launch_config_new_key] = data['Resources'].pop(
        manager_launch_config_orig_key)
    data['Resources']['ManagerAsg']['Properties']['LaunchConfigurationName']['Ref'] = manager_launch_config_new_key

    cmd_array = data['Resources'][manager_launch_config_new_key]['Properties']['UserData']['Fn::Base64']['Fn::Join'][1]
    insert_location = cmd_array.index(docker_launch_command)
    cmd_array = cmd_array[:insert_location] + new_content + cmd_array[insert_location:]
    data['Resources'][manager_launch_config_new_key]['Properties']['UserData']['Fn::Base64']['Fn::Join'][1] = cmd_array

    data['Resources'][node_launch_config_new_key] = data['Resources'].pop(
        node_launch_config_orig_key)
    data['Resources']['NodeAsg']['Properties']['LaunchConfigurationName']['Ref'] = node_launch_config_new_key

    cmd_array_node = data['Resources'][node_launch_config_new_key]['Properties']['UserData']['Fn::Base64']['Fn::Join'][1]
    insert_location_node = cmd_array_node.index(docker_launch_command)
    cmd_array_node = cmd_array_node[:insert_location_node] + new_content + cmd_array_node[insert_location_node:]
    data['Resources'][node_launch_config_new_key]['Properties']['UserData']['Fn::Base64']['Fn::Join'][1] = cmd_array_node

    cloudformation_template_name = u"{}.json".format(cfn_name)
    outdir = u"dist/aws/{}".format(release_channel)
    # if the directory doesn't exist, create it.
    if not os.path.exists(outdir):
        os.makedirs(outdir)

    outfile = u"{}/{}".format(outdir, cloudformation_template_name)

    with open(outfile, 'w') as outf:
        json.dump(data, outf, indent=4, sort_keys=True)

    # TODO: json validate

    print(u"DDC Dev Cloudformation template created in {}".format(outfile))
    return upload_cfn_template(release_channel, cloudformation_template_name, outfile, cfn_type="dev")

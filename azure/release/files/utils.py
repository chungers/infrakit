#!/usr/bin/env python
import os
import json
import time
from datetime import datetime
import boto
import urllib2
from boto import ec2
import re
from collections import (OrderedDict, namedtuple)

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

def upload_rg_template(release_channel, cloudformation_template_name, tempfile, cfn_type=''):

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

def create_rg_template(vhd_sku, vhd_version, release_channel, docker_version,
                        docker_for_azure_version, edition_version, cfn_template, cloudformation_template_name):
    # check if file exists before opening.
    flat_edition_version = edition_version.replace(" ", "").replace("_", "").replace("-", "")
    flat_edition_version_upper = flat_edition_version.capitalize()

    with open(cfn_template) as data_file:
        data = json.load(data_file)
    
    data['variables']['Description'] = u"Docker for Azure {0} ({1})".format(docker_version, edition_version)
    data['variables']['imageSku'] = vhd_sku
    data['variables']['imageVersion'] = vhd_version
    data['variables']['dockerForIAASVersion'] = docker_for_azure_version

    # Updated Manager custom data
    manager_data = buildCustomData('custom-data_manager.sh')
    data['variables']['customDataManager'] = '[concat(' + ', '.join(manager_data) + ')]'
    # Updated Worker custom data
    worker_data = buildCustomData('custom-data_worker.sh')
    data['variables']['customDataWorker'] = '[concat(' + ', '.join(worker_data) + ')]'


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

    return outfile

# @TODO VERIFY CLOUD TEMPLATE
# @TODO IMPLEMENT DDC TEMPLATE
def create_rg_cloud_template(release_channel, docker_version,
                        docker_for_azure_version, edition_version, cfn_template, cloudformation_template_name):
    with open(cfn_template) as data_file:
        data = json.load(data_file)

    # Updated Manager custom data
    manager_data = buildCustomData('custom-data_manager_cloud.sh')
    data['variables']['customDataManager'] = '[concat(' + ', '.join(manager_data) + ')]'

    parameters = data.get('parameters')
    if parameters:
        new_parameters = {
            "DockerCloudClusterName" : {
                "type": "string",
                "metadata": {
                    "description": "Name of the cluster (namespace/cluster_name) to be used when registering this Swarm with Docker Cloud\n\nMust be in the format 'namespace/cluster_name' and must only contain letters, digits and hyphens"
                }
            },
            "DockerCloudUsername" : {
                "type" : "string",
                "metadata": {
                    "description": "Docker ID username to use during registration\n\nMust only contain letters or digits"
                }
            },
            "DockerCloudAPIKey" : {
                "type" : "securestring",
                "metadata": {
                    "description": "Docker ID API key to use during registration"
                }
            }
        }
        parameters.update(new_parameters)
    
    variables = data.get('variables')
    if variables:
        new_variables = {
            "dockerCloudClusterName" : "[parameters('DockerCloudClusterName')]",
            "dockerCloudUsername" : "[parameters('DockerCloudUsername')]",
            "dockerCloudAPIKey" : "[parameters('DockerCloudAPIKey')]",
        }
        variables.update(new_variables)

    outputs = data.get('outputs')
    if outputs:
        new_outputs = {
            "ConnectToThisCluster": {
                "type": "string",
                "value": "[concat('docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_HOST dockercloud/client ', variables('DockerCloudClusterName'))]"
            }
        }
        outputs.update(new_outputs)
    
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

    return outfile

def create_rg_ddc_template(vhd_sku, vhd_version, release_channel, docker_version,
                        docker_for_azure_version, edition_version, cfn_template, cloudformation_template_name):
    with open(cfn_template) as data_file:
        data = json.load(data_file)

    # Updated Manager custom data
    manager_data = buildCustomData('custom-data_manager_ddc.sh')
    data['variables']['customDataManager'] = '[concat(' + ', '.join(manager_data) + ')]'
    # Updated Worker custom data
    worker_data = buildCustomData('custom-data_worker_ddc.sh')
    data['variables']['customDataWorker'] = '[concat(' + ', '.join(worker_data) + ')]'

    data['variables']['imageSku'] = vhd_sku
    data['variables']['imageVersion'] = vhd_version

    parameters = data.get('parameters')
    if parameters:
        new_parameters = {
            "DDCUsername": {
                "defaultValue": "admin",
                "type": "String",
                "metadata": {
                    "description": "Please enter the username you want to use for Docker Datacenter."
                }
            },
            "DDCPassword": {
                "minLength": 8,
                "maxLength": 40,
                "type": "SecureString",
                "metadata": {
                    "description": "Please enter the password you want to use for Docker Datacenter."
                }
            }
        }
        parameters.update(new_parameters)
    
    variables = data.get('variables')
    if variables:
        new_variables = {
            "ddcUser": "[parameters('DDCUsername')]",
            "ddcPass": "[parameters('DDCPassword')]"
        }
        variables.update(new_variables)
    
    for key, val in enumerate(data.get('resources')):
        if val['name'] == "[variables('managerNSGName')]":
            security_rules = val['properties']['securityRules']
            new_security_rule = {
                "name": "ddc",
                "properties": {
                    "description": "Allow UCP",
                    "protocol": "Tcp",
                    "sourcePortRange": "*",
                    "destinationPortRange": "443",
                    "sourceAddressPrefix": "*",
                    "destinationAddressPrefix": "*",
                    "access": "Allow",
                    "priority": 206,
                    "direction": "Inbound"
                }
            }
            security_rules.append(new_security_rule)
            break
    outputs = data.get('outputs')
    if outputs:
        new_outputs = {
            "DDCLoginURL": {
                "type": "String",
                "value": "[concat('https://', reference(resourceId('Microsoft.Network/publicIPAddresses', variables('lbSSHPublicIPAddressName'))).ipAddress)]"
            },
            "DDCUsername": {
                "type": "String",
                "value": "[variables('ddcUser')]"
            }
        }
        outputs.update(new_outputs)

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

    return outfile
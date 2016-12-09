#!/usr/bin/env python
import os
import json
import time
from datetime import datetime
import boto
import urllib2
from boto import ec2

NOW = datetime.now()
NOW_STRING = NOW.strftime("%m_%d_%Y")
REGIONS = ['us-west-1', 'us-west-2', 'us-east-1',
           'eu-west-1', 'eu-central-1', 'ap-southeast-1',
           'ap-northeast-1', 'ap-southeast-2', 'ap-northeast-2',
           'sa-east-1', 'ap-south-1', 'us-east-2', 'ca-central-1']
AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
S3_BUCKET_NAME = "docker-for-aws"


# file with list of aws account_ids
ACCOUNT_LIST_FILE_URL = u"https://s3.amazonaws.com/docker-for-aws/data/accounts.txt"
DOCKER_AWS_ACCOUNT_URL = "https://s3.amazonaws.com/docker-for-aws/data/docker_accounts.txt"
CS_AMI_LIST_PATH = u"data/ami/cs/{}/ami_list.json"


def str2bool(v):
    return v.lower() in ("yes", "true", "t", "1")


def get_ami(conn, ami_id):
    ''' Gets a single AMI as a boto.ec2.image.Image object '''

    attempts = 0
    max_attempts = 5
    while (attempts < max_attempts):
        try:
            attempts += 1
            images = conn.get_all_images(ami_id)
        except boto.exception.EC2ResponseError:
            msg = "Could not find AMI '%s' in region '%s'" % (ami_id, conn.region.name)
            if attempts < max_attempts:
                # The API call to initiate an AMI copy is not blocking, so the
                # copied AMI may not be available right away
                print(msg + ' so waiting 5 seconds and retrying')
                time.sleep(5)
            else:
                raise Exception(msg)

    if len(images) != 1:
        msg = "Somehow more than 1 AMI was detected - this is a weird error"
        raise Exception(msg)

    return images[0]


def wait_for_ami_to_be_available(conn, ami_id):
    ''' Blocking wait until the AMI is available '''

    ami = get_ami(conn, ami_id)

    while ami.state != 'available':
        print("{} in {} not available ({}), waiting 30 seconds ...".format(
            ami_id, conn.region.name, ami.state))
        time.sleep(30)
        ami = get_ami(conn, ami_id)

        if ami.state == 'failed':
            msg = "AMI '%s' is in a failed state and will never be available" % ami_id
            raise Exception(msg)

    return ami


def copy_amis(ami_id, ami_src_region, moby_image_name, moby_image_description, release_channel):
    """ Copy the given AMI to all target regions."""
    ami_list = {}
    for region in REGIONS:
        print(u"Copying AMI to {}".format(region))
        con = ec2.connect_to_region(region, aws_access_key_id=AWS_ACCESS_KEY_ID,
                                    aws_secret_access_key=AWS_SECRET_ACCESS_KEY)
        image = con.copy_image(ami_src_region, ami_id,
                               name=moby_image_name, description=moby_image_description)
        ami_list[region] = {"HVM64": image.image_id, "HVMG2": "NOT_SUPPORTED"}

        con.create_tags(image.image_id, {'channel': release_channel, 'date': NOW_STRING})

    return ami_list


def strip_and_padd(value):
    """ strip any extra characters and make sure to left pad
    with 0's until we get 12 characters"""
    if value:
        return value.strip().rjust(12, '0')
    return value


def get_account_list(account_file_url):
    """ Get the account list from a given url"""
    print(u"account_file_url = {}".format(account_file_url))
    response = urllib2.urlopen(account_file_url)
    return [strip_and_padd(line) for line in response.readlines() if len(line) > 0]


def get_ami_list(ami_list_url):
    """ Given a url to an ami json file, return the json object"""
    print(u"get_ami_list = {}".format(ami_list_url))
    response = urllib2.urlopen(ami_list_url)
    return json.load(response)


def approve_accounts(ami_list, account_list):
    """ TODO: What happens if one of the accounts isn't valid. Should we have an AccountID
        validation step?
        If one accountId isn't valid, the whole batch will get messed up.
    """
    slice_count = 100
    for ami in ami_list:
        print(u"Approve accounts for AMI: {}".format(ami))
        region = ami
        ami_id = ami_list.get(region).get('HVM64')
        print(u"Approve accounts for AMI: {} ; Region: {} ".format(ami_id, region))
        con = ec2.connect_to_region(region, aws_access_key_id=AWS_ACCESS_KEY_ID,
                                    aws_secret_access_key=AWS_SECRET_ACCESS_KEY)

        print("Wait until the AMI is available before we continue. We can't approve until then.")
        wait_for_ami_to_be_available(con, ami_id)

        if len(account_list) > slice_count:
            # if we have more then what we feel is enough account ids for one request
            # split them up into smaller requests (batches).
            count = 0
            print(u"Approving Accounts in batches")
            while count < len(account_list):
                print(u"[{}:{}] Approving batch {} - {}".format(
                    ami_id, region, count, count+slice_count))
                a_list = account_list[count:count+slice_count]
                con.modify_image_attribute(ami_id, operation='add',
                                           attribute='launchPermission', user_ids=a_list)
                count += slice_count
        else:
            print(u"[{}:{}] Approving All accounts at once".format(ami_id, region))
            con.modify_image_attribute(ami_id, operation='add',
                                       attribute='launchPermission', user_ids=account_list)

    return ami_list


def set_ami_public(ami_list):
    """ set the AMI's to public """
    for ami in ami_list:
        print(u"Set AMI public: {}".format(ami))
        region = ami
        ami_id = ami_list.get(region).get('HVM64')
        print(u"Set public AMI: {} ; Region: {} ".format(ami_id, region))
        con = ec2.connect_to_region(region, aws_access_key_id=AWS_ACCESS_KEY_ID,
                                    aws_secret_access_key=AWS_SECRET_ACCESS_KEY)

        print("Wait until the AMI is available before we continue.")
        wait_for_ami_to_be_available(con, ami_id)

        con.modify_image_attribute(ami_id,
                                   operation='add',
                                   attribute='launchPermission',
                                   groups='all')


def upload_cfn_template(release_channel, cloudformation_template_name, tempfile, cfn_type=''):

    # upload to s3, make public, return s3 URL
    s3_host_name = u"https://{}.s3.amazonaws.com".format(S3_BUCKET_NAME)
    s3_path = u"aws/{}/{}".format(release_channel, cloudformation_template_name)
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


def upload_ami_list(ami_list_json, docker_version):

    # upload to s3, make public, return s3 URL
    s3_host_name = u"https://{}.s3.amazonaws.com".format(S3_BUCKET_NAME)
    s3_path = u"data/cs_amis.json"
    s3_cs_path = CS_AMI_LIST_PATH.format(docker_version)
    s3_full_url = u"{}/{}".format(s3_host_name, s3_path)

    s3conn = boto.connect_s3(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(S3_BUCKET_NAME)

    print(u"Upload ami list json template to {} in {} s3 bucket".format(s3_path, S3_BUCKET_NAME))
    key = bucket.new_key(s3_path)
    key.set_metadata("Content-Type", "application/json")
    key.set_contents_from_string(ami_list_json)
    key.set_acl("public-read")

    print(u"Upload ami list json template to {} in {} s3 bucket".format(s3_cs_path, S3_BUCKET_NAME))
    key = bucket.new_key(s3_cs_path)
    key.set_metadata("Content-Type", "application/json")
    key.set_contents_from_string(ami_list_json)
    key.set_acl("public-read")

    return s3_full_url


def create_cfn_template(amis, release_channel, docker_version,
                        docker_for_aws_version, edition_version, cfn_template, cfn_name):
    # check if file exists before opening.

    flat_edition_version = edition_version.replace(" ", "").replace("_", "").replace("-", "")
    flat_edition_version_upper = flat_edition_version.capitalize()

    with open(cfn_template) as data_file:
        data = json.load(data_file)

    data['Description'] = u"Docker for AWS {0} ({1})".format(docker_version, edition_version)
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

    data['Resources'][manager_launch_config_new_key] = data['Resources'].pop(
        manager_launch_config_orig_key)
    data['Resources']['ManagerAsg']['Properties']['LaunchConfigurationName']['Ref'] = manager_launch_config_new_key

    data['Resources'][node_launch_config_new_key] = data['Resources'].pop(
        node_launch_config_orig_key)
    data['Resources']['NodeAsg']['Properties']['LaunchConfigurationName']['Ref'] = node_launch_config_new_key

    cloudformation_template_name = u"{}.json".format(cfn_name)
    outdir = u"dist/aws/{}".format(release_channel)
    # if the directory doesn't exist, create it.
    if not os.path.exists(outdir):
        os.makedirs(outdir)

    outfile = u"{}/{}".format(outdir, cloudformation_template_name)

    with open(outfile, 'w') as outf:
        json.dump(data, outf, indent=4, sort_keys=True)

    # TODO: json validate

    print(u"Cloudformation template created in {}".format(outfile))
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

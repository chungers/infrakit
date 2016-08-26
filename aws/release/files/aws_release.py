#!/usr/bin/env python

import os
import json
import argparse
import time
from datetime import datetime
import boto
from boto import ec2

NOW = datetime.now()
NOW_STRING = NOW.strftime("%m_%d_%Y")
REGIONS = ['us-west-1', 'us-west-2', 'us-east-1',
           'eu-west-1', 'eu-central-1', 'ap-southeast-1',
           'ap-northeast-1', 'ap-southeast-2', 'ap-northeast-2',
           'sa-east-1']
AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
CFN_TEMPLATE = '/home/docker/docker_for_aws.template'
CFN_DDC_TEMPLATE = '/home/docker/docker_for_aws_ddc.template'

# file with list of aws account_ids
ACCOUNT_LIST_FILE_URL = u"https://s3.amazonaws.com/docker-for-aws/data/accounts.txt"


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


def upload_cfn_template(release_channel, cloudformation_template_name, tempfile):

    # upload to s3, make public, return s3 URL
    s3_bucket_name = "docker-for-aws"
    s3_host_name = u"https://{}.s3.amazonaws.com".format(s3_bucket_name)
    s3_path = u"aws/{}/{}".format(release_channel, cloudformation_template_name)
    s3_path_latest = u"aws/{}/latest.json".format(release_channel)
    s3_full_url = u"{}/{}".format(s3_host_name, s3_path)

    s3conn = boto.connect_s3(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(s3_bucket_name)

    print(u"Upload Cloudformation template to {} in {} s3 bucket".format(s3_path, s3_bucket_name))
    key = bucket.new_key(s3_path)
    key.set_metadata("Content-Type", "application/json")
    key.set_contents_from_filename(tempfile)
    key.set_acl("public-read")

    if release_channel == 'nightly':
        print("This is a nightly build, update the latest.json file.")
        print(u"Upload Cloudformation template to {} in {} s3 bucket".format(
            s3_path_latest, s3_bucket_name))
        key = bucket.new_key(s3_path_latest)
        key.set_metadata("Content-Type", "application/json")
        key.set_contents_from_filename(tempfile)
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


def copy_amis(ami_id, ami_src_region, moby_image_name, moby_image_description, release_channel):
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
    print(u"account_file_url = {}".format(account_file_url))
    import urllib2
    response = urllib2.urlopen(account_file_url)
    return [strip_and_padd(line) for line in response.readlines() if len(line) > 0]


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


def main():
    parser = argparse.ArgumentParser(description='Release Docker for AWS')
    parser.add_argument('-d', '--docker_version',
                        dest='docker_version', required=True,
                        help="Docker version (i.e. 1.12.0-rc4)")
    parser.add_argument('-e', '--edition_version',
                        dest='edition_version', required=True,
                        help="Edition version (i.e. Beta 4)")
    parser.add_argument('-a', '--ami_id',
                        dest='ami_id', required=True,
                        help="ami-id for the Moby AMI we just built (i.e. ami-123456)")
    parser.add_argument('-s', '--ami_src_region',
                        dest='ami_src_region', required=True,
                        help="The reason the source AMI was built in (i.e. us-east-1)")
    parser.add_argument('-c', '--channel',
                        dest='channel', default="beta",
                        help="release channel (beta, alpha, rc, nightly)")
    parser.add_argument('-u', '--channel_ddc',
                        dest='channel_ddc', default="alpha",
                        help="DDC release channel (beta, alpha, rc, nightly)")                    
    parser.add_argument('-l', '--account_list_url',
                        dest='account_list_url', default=ACCOUNT_LIST_FILE_URL,
                        help="The URL for the aws account list for ami approvals")

    args = parser.parse_args()

    # TODO: wait for source AMI to be ready?
    # https://github.com/Answers4AWS/distami/blob/master/distami/utils.py#L87

    # TODO: Don't hard code beta
    release_channel = args.channel
    release_ddc_channel = args.channel_ddc
    docker_version = args.docker_version
    # TODO change to something else? where to get moby version?
    moby_version = docker_version
    edition_version = args.edition_version
    flat_edition_version = edition_version.replace(" ", "")
    docker_for_aws_version = u"aws-v{}-{}".format(docker_version, flat_edition_version)
    docker_for_aws_ddc_version = u"aws-v{}-{}-ddc-tp1".format(docker_version, flat_edition_version)
    IMAGE_NAME = u"Moby Linux {}".format(docker_for_aws_version)
    IMAGE_DESCRIPTION = u"The best OS for running Docker, version {}".format(moby_version)
    print("\n Variables")
    print(u"release_channel={}".format(release_channel))
    print(u"docker_version={}".format(docker_version))
    print(u"edition_version={}".format(edition_version))
    print(u"ami_id={}".format(args.ami_id))
    print(u"ami_src_region={}".format(args.ami_src_region))
    print(u"account_list_url={}".format(args.account_list_url))
    if not args.account_list_url:
        print("account_list_url parameter is None, defaulting")
        account_list_url = ACCOUNT_LIST_FILE_URL
    else:
        account_list_url = args.account_list_url
    print(u"account_list_url={}".format(account_list_url))

    print("Copy AMI to each region..")
    ami_list = copy_amis(args.ami_id, args.ami_src_region,
                         IMAGE_NAME, IMAGE_DESCRIPTION, release_channel)
    ami_list_json = json.dumps(ami_list, indent=4, sort_keys=True)
    print(u"AMI copy complete. AMI List: \n{}".format(ami_list_json))
    print("Get account list..")
    account_list = get_account_list(account_list_url)
    print(u"Approving AMIs for {} accounts..".format(len(account_list)))
    approve_accounts(ami_list, account_list)
    print("Accounts have been approved.")
    print("Create CloudFormation template..")
    s3_url = create_cfn_template(ami_list, release_channel, docker_version,
                                 docker_for_aws_version, edition_version, CFN_TEMPLATE, docker_for_aws_version)
    print("Create DDC CloudFormation template..")
    s3_ddc_url = create_cfn_template(ami_list, release_ddc_channel, docker_version,
                                 docker_for_aws_version, edition_version, CFN_DDC_TEMPLATE, docker_for_aws_ddc_version)

    # TODO: git commit, tag release. requires github keys, etc.

    print("------------------")
    print(u"Finshed.. CloudFormation URL={0} \n \t DDC_URL={1} \n".format(s3_url, s3_ddc_url))
    print("Don't forget to tag the code (git tag -a v{0}-{1} -m {0}; git push --tags)".format(
        docker_version, flat_edition_version))
    print("------------------")


if __name__ == '__main__':
    main()

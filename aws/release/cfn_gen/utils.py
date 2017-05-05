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
           'eu-west-1', 'eu-west-2', 'eu-central-1', 'ap-southeast-1',
           'ap-northeast-1', 'ap-southeast-2', 'ap-northeast-2',
           'sa-east-1', 'ap-south-1', 'us-east-2', 'ca-central-1']
AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
S3_BUCKET_NAME = os.getenv('S3_BUCKET', 'docker-ci-editions')
CFN_S3_BUCKET_NAME = os.getenv('UPLOAD_S3_BUCKET', 'docker-for-aws')
CFN_AWS_ACCESS_KEY_ID = os.getenv('UPLOAD_S3_KEY', AWS_ACCESS_KEY_ID)
CFN_AWS_SECRET_ACCESS_KEY = os.getenv('UPLOAD_S3_SECRET', AWS_SECRET_ACCESS_KEY)
MOBY_COMMIT = os.getenv('MOBY_COMMIT',"unknown-moby-commit")


# file with list of aws account_ids
ACCOUNT_LIST_FILE_URL = u"https://s3.amazonaws.com/docker-for-aws/data/accounts.txt"
DOCKER_AWS_ACCOUNT_URL = "https://s3.amazonaws.com/docker-for-aws/data/docker_accounts.txt"
AMI_LIST_PATH = u"data/ami/{}/ami_list.json"


def str2bool(v):
    return v.lower() in ("yes", "true", "t", "1")

def is_json(cfn_template):
  try:
    with open(cfn_template) as data_file:
      json_object = json.load(data_file)
  except ValueError, e:
    return False
  return True

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

def upload_ami_list(ami_list_json, editions_version, docker_version):
	
    # upload to s3, make public, return s3 URL
    s3_host_name = u"https://{}.s3.amazonaws.com".format(S3_BUCKET_NAME)
    s3_path = u"ami/{}/ami_list.json".format(MOBY_COMMIT)
    s3_full_url = u"{}/{}".format(s3_host_name, s3_path)
    print(u"Upload ami list json template to {} in {} s3 bucket".format(s3_path, S3_BUCKET_NAME))
    s3conn = boto.connect_s3(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(S3_BUCKET_NAME)

    key = bucket.new_key(s3_path)
    key.set_metadata("Content-Type", "application/json")
    key.set_metadata("editions_version", editions_version)
    key.set_metadata("docker_version", docker_version)
    key.set_contents_from_string(ami_list_json)
    key.set_acl("public-read")

    return s3_full_url

def upload_cfn_template(release_channel, cloudformation_template_name,
                        tempfile, cfn_type=None):

    # upload to s3, make public, return s3 URL
    s3_host_name = u"https://{}.s3.amazonaws.com".format(CFN_S3_BUCKET_NAME)
    channel = release_channel[:release_channel.find("-")]
    if MOBY_COMMIT:
        channel = u"{}/{}".format(channel, MOBY_COMMIT)
    s3_path = u"aws/{}/{}.json".format(channel, u"Docker{}".format(cloudformation_template_name[cloudformation_template_name.find("-aws"):]))
    latest_name = "latest.json"
    if cfn_type:
        latest_name = "{}-latest.json".format(cfn_type)

    s3_path_latest = u"aws/{}/{}".format(release_channel, latest_name)
    s3_full_url = u"{}/{}".format(s3_host_name, s3_path)

    s3conn = boto.connect_s3(CFN_AWS_ACCESS_KEY_ID, CFN_AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(CFN_S3_BUCKET_NAME)

    print(u"Upload Cloudformation template to {} in {} s3 bucket".format(s3_path, CFN_S3_BUCKET_NAME))
    key = bucket.new_key(s3_path)
    key.set_metadata("Content-Type", "application/json")
    key.set_contents_from_filename(tempfile)
    key.set_acl("public-read")

    if release_channel == 'nightly' or release_channel == 'cloud-nightly':
        print("This is a nightly build, update the latest.json file.")
        print(u"Upload Cloudformation template to {} in {} s3 bucket".format(
            s3_path_latest, CFN_S3_BUCKET_NAME))
        key = bucket.new_key(s3_path_latest)
        key.set_metadata("Content-Type", "application/json")
        key.set_contents_from_filename(tempfile)
        key.set_acl("public-read")

    return s3_full_url


def publish_cfn_template(release_channel, docker_for_aws_version):
    # upload to s3, make public, return s3 URL
    s3_host_name = u"https://{}.s3.amazonaws.com".format(CFN_S3_BUCKET_NAME)
    s3_path = u"aws/{}/{}.json".format(release_channel, docker_for_aws_version)

    print(u"Update the latest.json file to the release of {} in {} channel.".format(docker_for_aws_version, release_channel))
    latest_name = "latest.json"
    s3_path_latest = u"aws/{}/{}".format(release_channel, latest_name)
    s3_full_url = u"{}/{}".format(s3_host_name, s3_path_latest)

    s3conn = boto.connect_s3(CFN_AWS_ACCESS_KEY_ID, CFN_AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(CFN_S3_BUCKET_NAME)

    print(u"Copy Cloudformation template from {} to {} s3 bucket".format(s3_path, s3_path_latest))
    srckey = bucket.get_key(s3_path)
    dstkey = bucket.new_key(s3_path_latest)
    srckey.copy(CFN_S3_BUCKET_NAME, dstkey, preserve_acl=True, validate_dst_bucket=True)
    return s3_full_url


def create_cfn_template(template_class, amis, release_channel,
                        docker_version, docker_for_aws_version,
                        edition_addon, cfn_name,
                        cfn_type=None):

    cloudformation_template_name = u"{}.json".format(cfn_name)
    curr_path = os.path.dirname(__file__)

    outdir = u'dist/aws/{}'.format(release_channel)
    # if the directory doesn't exist, create it.
    if not os.path.exists(outdir):
        os.makedirs(outdir)
    outfile = u"{}/{}".format(outdir, cloudformation_template_name)

    aws_template = template_class(
        docker_version, docker_for_aws_version,
        edition_addon, release_channel, amis)
    aws_template.build()

    new_template = json.loads(aws_template.generate_template())

    with open(outfile, 'w') as newfile:
        json.dump(new_template, newfile, sort_keys=True, indent=4, separators=(',', ': '))

    # TODO: json validate
    if is_json(outfile) is False:
      print(u"ERROR: Cloudformation template invalid in {}".format(outfile))
    return outfile

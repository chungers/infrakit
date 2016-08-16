#!/usr/bin/env python

import os
from boto import ec2
from boto import cloudformation
from boto.s3.connection import S3Connection
from datetime import datetime, timedelta

AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
s3_bucket_name = "docker-for-aws"

REGIONS = ['us-west-1', 'us-west-2', 'us-east-1',
           'eu-west-1', 'eu-central-1', 'ap-southeast-1',
           'ap-northeast-1', 'ap-southeast-2', 'ap-northeast-2',
           'sa-east-1']

EXPIRE_AGE = 30
CFN_EXPIRE_AGE = 2
NOW = datetime.now()
EXPIRE_DATE = NOW - timedelta(EXPIRE_AGE)
CFN_EXPIRE_DATE = NOW - timedelta(CFN_EXPIRE_AGE)

print("Cleanup AMIs and Snapshots")

for region in REGIONS:
    print(u"Region {}".format(region))
    con = ec2.connect_to_region(region, aws_access_key_id=AWS_ACCESS_KEY_ID,
                                aws_secret_access_key=AWS_SECRET_ACCESS_KEY)

    images = con.get_all_images(owners=['self'], filters={'tag:channel': 'nightly'})
    for image in images:
	print(u"{}: {}".format(image.name, image.tags))
        image_date_tag = image.tags.get('date')
        image_date = datetime.strptime(image_date_tag, "%m_%d_%Y")
	if image_date < EXPIRE_DATE:
            print("Too old, cleanup")
       	    image.deregister(delete_snapshot=True)
            print("Image is cleaned up")
        else:
            print("Still good, keep it around.")
    print("----------")

print("Clean up CloudFormation Files")

conn = S3Connection(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
bucket = conn.get_bucket(s3_bucket_name)
files = list(bucket.list("aws/nightly/"))

for key in files:
    # we only care about json files.
    if not key.name.endswith(".json"):
        continue

    key_date = datetime.strptime(key.last_modified, '%Y-%m-%dT%H:%M:%S.000Z')
    if key_date < EXPIRE_DATE:
        print(u"{} is {}, which is too old (< {}), remove it.".format(key.name, key.last_modified, EXPIRE_DATE))
	key.delete()

print("Clean up any left over CFN stacks")
for region in REGIONS:
    print(u"region={}".format(region))
    connection = cloudformation.connect_to_region(region)
    stacks = connection.describe_stacks()
    for stack in stacks:
        print(stack.stack_name)
        if stack.tags.get('channel') == 'nightly':
            cfn_date_tag = stack.tags.get('date')
            cfn_date = datetime.strptime(image_date_tag, "%m_%d_%Y")
            if cfn_date < CFN_EXPIRE_DATE:
                print("Too old, cleanup")
                connection.delete_stack(stack.stack_id)
                print("cfn stack is cleaned up")
            else:
                print("Still good, keep it around.")
        else:
            print("Not nightly CFN, ignore.")

print("Finished")

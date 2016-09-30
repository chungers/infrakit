#!/usr/bin/env python
# build a CS AMI.
import json
import argparse
import sys
import boto
from os import path
sys.path.append( path.dirname( path.dirname( path.abspath(__file__) ) ) )


from utils import (S3_BUCKET_NAME, str2bool, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)


def main():
    print("------------------")
    print(u"update s3 with cs AMI")
    print("------------------")
    parser = argparse.ArgumentParser(description='CS AMI Copy and approve')
    parser.register('type','bool', str2bool)
    parser.add_argument('-d', '--docker_version',
                        dest='docker_version', required=True,
                        help="Docker version (i.e. 1.12.0-rc4)")
    parser.add_argument('-a', '--ami_id',
                        dest='ami_id', required=True,
                        help="ami-id for the Moby AMI we just built (i.e. ami-123456)")
    parser.add_argument('-s', '--ami_src_region',
                        dest='ami_src_region', required=True,
                        help="The reason the source AMI was built in (i.e. us-east-1)")
    parser.add_argument('-u', '--update-latest', type='bool',
                        dest='update_latest', default=False)
    args = parser.parse_args()

    docker_version = args.docker_version

    docker_for_aws_version = u"aws-v{}".format(docker_version)
    print("\n Variables")
    print(u"docker_version={}".format(docker_version))
    print(u"ami_id={}".format(args.ami_id))
    print(u"ami_src_region={}".format(args.ami_src_region))

    # upload to s3, make public, return s3 URL
    root_path = u"data/ami/cs"
    latest_s3_path = u"{}/latest.txt".format(root_path)
    s3_root_path = u"{}/{}".format(root_path, docker_version)
    ami_s3_path = u"{}/ami.txt".format(s3_root_path)
    dv_s3_path = u"{}/docker_version.txt".format(s3_root_path)

    s3conn = boto.connect_s3(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(S3_BUCKET_NAME)

    ami_text = u"{},{}".format(args.ami_id, args.ami_src_region)

    print(u"Upload ami list json template to {} in {} s3 bucket".format(ami_s3_path, S3_BUCKET_NAME))
    key = bucket.new_key(ami_s3_path)
    key.set_metadata("Content-Type", "text/plain")
    key.set_contents_from_string(ami_text)
    key.set_acl("public-read")

    print(u"Upload docker_version to {} in {} s3 bucket".format(dv_s3_path, S3_BUCKET_NAME))
    key = bucket.new_key(dv_s3_path)
    key.set_metadata("Content-Type", "text/plain")
    key.set_contents_from_string(docker_version)
    key.set_acl("public-read")

    if args.update_latest:
        print(u"Upload latest {} in {} s3 bucket".format(latest_s3_path, S3_BUCKET_NAME))
        key = bucket.new_key(latest_s3_path)
        key.set_metadata("Content-Type", "text/plain")
        key.set_contents_from_string(docker_version)
        key.set_acl("public-read")

    print("------------------")
    print(u"Finshed updating s3 with cs AMI")
    print(u"AMIID : {}".format(args.ami_id))
    print(u"AWS Region : {}".format(args.ami_src_region))
    print(u"docker_version : {}".format(docker_version))
    print("------------------")


if __name__ == '__main__':
    main()

#!/usr/bin/env python

import os
from boto.s3.connection import S3Connection
from datetime import datetime

AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
s3_bucket_name = "docker-for-aws"

conn = S3Connection(AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
bucket = conn.get_bucket(s3_bucket_name)
files = list(bucket.list("aws/nightly/"))
files.reverse()

css = '<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous"><link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">'

html = "<!DOCTYPE html><html lang='en'><head><meta charset='utf-8'><meta http-equiv='X-UA-Compatible' content='IE=edge'><meta name='viewport' content='width=device-width, initial-scale=1'><title>Docker for AWS nightly buiilds</title>{}</head><body><div class='container'><h1>Docker for AWS nightly builds</h1><table class='table'><tr><th>Day</th><th>Last modified</th><th>Launch</th></tr>".format(css)
button = "<a href='https://console.aws.amazon.com/cloudformation/home?#/stacks/new?stackName=D4A-Nightly&templateURL={}' target='_blank'><img src='https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png'></a>"

for key in files:
    # we only care about json files.
    if not key.name.endswith(".json"):
	continue
    row = "<tr><td>{}</td><td>{}</td><td>{}</td></tr>"
    cfn_name = key.name.split("/")[-1]

    html += row.format(cfn_name, key.last_modified, button.format(key.generate_url(expires_in=0, query_auth=False)))


now = datetime.now()

html += "</table><p>last updated: {}</p></div></body></html>".format(now)

s3_path = "aws/nightly/index.html"
key = bucket.new_key(s3_path)
key.set_metadata("Content-Type", "text/html")
key.set_contents_from_string(html)
key.set_acl("public-read")


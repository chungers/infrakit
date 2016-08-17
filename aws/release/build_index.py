#!/usr/bin/env python

import os
from boto.s3.connection import S3Connection
from datetime import datetime
import json

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
    row = "<tr><td><a href='{0}'>{1}</a></td><td>{2}</td><td>{3}</td></tr>"
    cfn_name = key.name.split("/")[-1]
    url = key.generate_url(expires_in=0, query_auth=False)
    html += row.format(url, cfn_name, key.last_modified, button.format(url))


now = datetime.now()

html += "</table><h2>Latest Test Results</h2><table class='table table-bordered'><tr><th>Region</th><th>Test time in seconds</th><th>Result</th></tr>"

# test results
file_date = now.strftime("%m_%d_%Y")
results_file = "/home/ubuntu/out/{}_results.json".format(file_date)
if os.path.exists(results_file):
    with open(results_file) as data_file:
        data = json.load(data_file)

    for key, value in data.iteritems():
        status = value.get('status')
        if status == 'CREATE_COMPLETE':
		stack_status = "<span class='label label-success'>{}</span>".format(status)
        else:
		stack_status = "<span class='label label-danger'>{}</span>".format(status)
        html += "<tr><td>{}</td><td>{}</td><td>{}</td></tr>".format(
            key, value.get('total_time_secs'), stack_status)

else:
    html += "<tr><td>Not available</td></tr>"


html += "</table><p>last updated: {}</p></div></body></html>".format(now)

s3_path = "aws/nightly/index.html"
key = bucket.new_key(s3_path)
key.set_metadata("Content-Type", "text/html")
key.set_contents_from_string(html)
key.set_acl("public-read")

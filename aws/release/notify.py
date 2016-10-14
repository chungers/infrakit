#!/usr/bin/env python

import os
import requests
from datetime import datetime
import json


SLACK_INCOMING_WEB_HOOK = "https://hooks.slack.com/services/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"

payload = {
    "text": "Docker AWS Nightly Build Results",
    "channel": "#editions-dev",
    "attachments": []
}

def results(title, results_file, s3_path):

    now = datetime.now()
    full_date = now.strftime("%B %d, %Y")
    color = "good"
    full_title = "{} Nightly Build Results".format(title)
    attachment = {
        "fallback": full_title,
        "text": full_title,
        "title": "{} {}".format(full_title, full_date),
        "title_link": "https://docker-for-aws.s3.amazonaws.com/{}/index.html".format(s3_path),
        "color": color,
        "fields": []
    }

    # test results
    file_date = now.strftime("%m_%d_%Y")
    results_file = "/home/ubuntu/out/{}_{}.json".format(file_date, results_file)
    fields = []
    if os.path.exists(results_file):
        with open(results_file) as data_file:
            data = json.load(data_file)

        for key, value in data.iteritems():
            status = value.get('status')
            if status == 'CREATE_COMPLETE':
                stack_status = "success"
            else:
                color = 'danger'
                stack_status = status
            fields.append({"title": key, "value": stack_status, "short": True})

    attachment['fields'] = fields
    attachment['color'] = color
    return attachment


attachment_oss = results("Docker AWS", "results", "aws/nightly")
attachment_ddc = results("Docker AWS + DDC", "ddc_results", "aws/ddc-nightly")
attachment_cloud = results("Docker AWS + Cloud Federation", "cloud_results", "aws/cloud-nightly")
payload['attachments'] = [attachment_oss, attachment_ddc, attachment_cloud]

# send message
requests.post(SLACK_INCOMING_WEB_HOOK,
              json.dumps(payload), headers={'content-type': 'application/json'})

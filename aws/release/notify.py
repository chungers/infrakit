#!/usr/bin/env python

import os
import requests
from datetime import datetime
import json

now = datetime.now()
full_date = now.strftime("%B %d, %Y")
SLACK_INCOMING_WEB_HOOK = "https://hooks.slack.com/services/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"
color = "good"
payload = {
    "text": "Docker AWS Nightly Build Results",
    "channel": "#editions-dev",
    "attachments": []
}

attachment = {
    "fallback": "Docker AWS Nightly Build Results",
    "text": "Docker AWS Nightly Build Results",
    "title": full_date,
    "title_link": "https://docker-for-aws.s3.amazonaws.com/aws/nightly/index.html",
    "color": color,
    "fields": []
}


# test results
file_date = now.strftime("%m_%d_%Y")
results_file = "/home/ubuntu/out/{}_results.json".format(file_date)
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

payload['attachments'] = [attachment, ]

# send message
requests.post(SLACK_INCOMING_WEB_HOOK,
              json.dumps(payload), headers={'content-type': 'application/json'})

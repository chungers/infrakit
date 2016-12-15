import os
from boto import cloudformation
import requests
import json

AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')

REGIONS = ['us-west-1', 'us-west-2', 'us-east-1',
           'eu-west-1', 'eu-west-2', 'eu-central-1', 'ap-southeast-1',
           'ap-northeast-1', 'ap-southeast-2', 'ap-northeast-2',
           'sa-east-1', 'ap-south-1', 'us-east-2', 'ca-central-1']


def find_stacks():
    running_stacks = {}

    # get the list of running stacks on all regions.
    for region in REGIONS:
        print(u"region={}".format(region))
        stack_names = []
        connection = cloudformation.connect_to_region(region)
        stacks = connection.describe_stacks()
        for stack in stacks:
            print(stack.stack_name)
            stack_names.append(stack.stack_name)

        if len(stack_names) > 0:
            running_stacks[region] = stack_names

    return running_stacks


SLACK_INCOMING_WEB_HOOK = "https://hooks.slack.com/services/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"

payload = {
    "text": "Nightly Stack Audit: Currently running stacks",
    "channel": "#editions-dev",
    "attachments": []
}


def results(title, running_stacks):
    fields = []

    for region, stacks in running_stacks.iteritems():
        stack_str = ", ".join(stacks)
        fields.append({"title": region, "value": stack_str, "short": False})

    attachment = {
        "fallback": title,
        "text": "Here is the list of stacks that are running. If they are yours and you don't need them anymore, please delete.",
        "title": title,
        "color": "green",
        "fields": fields
    }

    return attachment

# find the stacks that are running.
running_stacks = find_stacks()
print(running_stacks)

# send notification to slack.
the_results = results("Currently Running Stacks on AWS", running_stacks)
payload['attachments'] = [the_results]

# send message
response = requests.post(SLACK_INCOMING_WEB_HOOK,
                         json.dumps(payload), headers={'content-type': 'application/json'})

print(response)
print(response.text)
print("done")

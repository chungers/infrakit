#!/usr/bin/env python
import argparse
import copy
from boto import cloudformation
import time
import json
from datetime import datetime

NOW = datetime.now()
NOW_STRING = NOW.strftime("%m_%d_%Y")
# Take in a CFN URL as a paramter and then try and start the CFN template in each region

PARAMETERS = [('ClusterSize', 2),
              ('InstanceType', 't2.micro'),
              ('KeyName', 'ken_cochrane'),
              ('ManagerInstanceType', 't2.micro'),
              ('ManagerSize', 3)]

DDC_PARAMETERS = [('ClusterSize', 2),
              ('InstanceType', 't2.micro'),
              ('KeyName', 'ken_cochrane'),
              ('ManagerInstanceType', 'm3.medium'),
              ('ManagerSize', 3),
              ('DDCPasswordSet', 'password'),
	      ('License', '{}')]

CLOUD_PARAMETERS = [('ClusterSize', 2),
              ('InstanceType', 't2.micro'),
              ('KeyName', 'ken_cochrane'),
              ('ManagerInstanceType', 't2.medium'),
              ('ManagerSize', 3),
              ('DockerCloudClusterName', 'daodoo/swarm-{}'.format(NOW.strftime("%m-%d-%Y"))),
              ('DockerCloudUsername', 'appaws'),
              ('DockerCloudAPIKey', '01b6eb3a-f5aa-414a-bfb0-4273819299f4')]

REGIONS = ['us-west-1', 'us-west-2', 'us-east-1',
           'eu-west-1', 'eu-central-1', 'ap-southeast-1',
           'ap-northeast-1', 'ap-southeast-2', 'ap-northeast-2',
           'sa-east-1', 'ap-south-1', 'us-east-2']

VALID_STACK_STATUSES = ['CREATE_IN_PROGRESS', 'CREATE_FAILED', 'CREATE_COMPLETE', 'ROLLBACK_IN_PROGRESS',
                        'ROLLBACK_FAILED', 'ROLLBACK_COMPLETE', 'DELETE_IN_PROGRESS', 'DELETE_FAILED',
                        'DELETE_COMPLETE', 'UPDATE_IN_PROGRESS', 'UPDATE_COMPLETE_CLEANUP_IN_PROGRESS',
                        'UPDATE_COMPLETE', 'UPDATE_ROLLBACK_IN_PROGRESS', 'UPDATE_ROLLBACK_FAILED',
                        'UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS', 'UPDATE_ROLLBACK_COMPLETE']

STACK_COMPLETE_STATUSES = ['CREATE_COMPLETE', 'CREATE_FAILED',
                           'ROLLBACK_FAILED', 'ROLLBACK_COMPLETE']


def check_stack_statuses(stacks):
    running_queue = copy.deepcopy(stacks)
    timeout = 90  # max runtime is 45 minutes total
    count = 0
    while len(running_queue) > 0:
        print("Sleeping for 30 seconds.")
        time.sleep(30)
        for key in running_queue.keys():
            region = key
            value = running_queue.get(key)
            connection = cloudformation.connect_to_region(region)
            stack_id = value.get('stack_id')
            stack = connection.describe_stacks(stack_id)[0]
            status = stack.stack_status
            print(u"{}:{}".format(region, status))
            if status in STACK_COMPLETE_STATUSES:
                # stack is done, save results and cleanup.
                print(u"{} is done, remove.".format(region, status))
                stop = time.time()
                value['stop_time'] = stop
                total_time = stop - value.get('start_time')
                value['total_time_secs'] = total_time
                value['status'] = status
                stacks[region] = value

                # remove from running queue
                running_queue.pop(region)
                for output in stack.outputs:
                    print('%s=%s (%s)' % (output.key, output.value, output.description))

                # if the stack failed, print the events to see what happened.
                if status in ['ROLLBACK_COMPLETE', 'ROLLBACK_FAILED']:
                    print("Stack failed, print events")
                    events = connection.describe_stack_events(stack_id)
                    for event in events:
                        print(u"  {} : {}".format(event, event.resource_status_reason))
                print("Deleting stack..")
                connection.delete_stack(stack_id)

        count += 1
        if count > timeout:
            print("Took too long, timing out")
            break

    return stacks


def run_cfn(connection, cloud_formation_url, stack_params, channel, name):
    stack_id = connection.create_stack(name,
                                       template_url=cloud_formation_url,
                                       parameters=stack_params,
                                       timeout_in_minutes=30,
                                       tags={'channel': channel, 'date': NOW_STRING},
                                       capabilities=['CAPABILITY_IAM'])
    print(stack_id)
    return stack_id


def run_stacks(cloud_formation_url, stack_params, channel, name):

    stacks = {}
    for region in REGIONS:
        print(u"Create Stack for {}".format(region))
        connection = cloudformation.connect_to_region(region)
        start = time.time()
        stack_id = run_cfn(connection, cloud_formation_url, stack_params, channel, name)
        stacks[region] = {
            'stack_id': stack_id,
            'start_time': start,
        }
    return stacks


def main():
    print("Start")
    parser = argparse.ArgumentParser(description='Release Docker for AWS')
    parser.add_argument('-c', '--cloud_formation_url',
                        dest='cloud_formation_url', required=True,
                        help="The CloudFormtaion URL to test")
    parser.add_argument('-f', '--results_file_name',
                        dest='results_file_name', required=True,
                        help="The Results file name")
    parser.add_argument('-t', '--stack_type',
                        dest='stack_type', required=True,
                        default="oss",
                        help="The type of stack (oss, ddc, cloud)")
    args = parser.parse_args()

    if not args.stack_type or args.stack_type.lower() == 'oss':
        stack_params = PARAMETERS
        channel = "nightly"
        name = u"Nite-{}".format(NOW.strftime("%m%d%Y%f")[:12])
    elif args.stack_type.lower() == 'cloud':
        stack_params = CLOUD_PARAMETERS
        channel = "cloud-nightly"
        name = u"Cloud-Nite-{}".format(NOW.strftime("%m%d%Y%f")[:12])
    else:
        stack_params = DDC_PARAMETERS
        channel = "ddc-nightly"
        name = u"DDC-Nite-{}".format(NOW.strftime("%m%d%Y%f")[:12])

    print(u"Channel: {}".format(channel))
    print(u"Name: {}".format(name))
    stacks = run_stacks(args.cloud_formation_url, stack_params, channel, name)
    results_file = args.results_file_name
    print(stacks)
    results = check_stack_statuses(stacks)
    print(results)

    outfile = u"{}/{}".format("/home/ubuntu/out", u"{}_{}.json".format(NOW_STRING, results_file))
    print(u"Outfile = {}".format(outfile))
    with open(outfile, 'w') as outf:
        json.dump(results, outf, indent=4, sort_keys=True)
    print("Done")

if __name__ == '__main__':
    main()

#!/usr/bin/env python

import os
import re
from boto import cloudformation, logs

AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')


# Gets all the available aws regions and deletes
# the regions which are special cases from this list
def get_regions():
    regions = cloudformation.regions()

    regions_list = []
    for region in regions:
        regions_list.append(re.findall('(?<=RegionInfo:).*', str(region))[0])

    regions_list.remove("cn-north-1")
    regions_list.remove("us-gov-west-1")

    return regions_list


# record of all deleted stacks
stack_dict = {}

# Goal is to delete all CloudWatch log groups that are
# no longer needed this is defined as log groups with
# nite in their name, or log groups whose stack has
# been deleted. Deleted stacks are either in the stack
# list or not there if they have been deleted more than
# 90 days ago
for region in get_regions():
    print(region)

    connection_stack = cloudformation.connect_to_region(region)
    stacks = connection_stack.list_stacks()

    # stack is a StackSummary object
    # Stack summary object has the following attributes
    # connection, creation_time, description, disable_rollback,
    # notification_arns, outputs, parameters, tags, stack_id, stack_status,
    # stack_status_reason, stack_name

    for stack in stacks:
        stack_dict[stack.stack_name] = stack

    connection_logs = logs.connect_to_region(region)

    log_dict = connection_logs.describe_log_groups()
    next_token_log = -1
    while next_token_log is not None:
        # Use stack name from stack and see if matches to a portion of log
        # group name use (?<=aws/lambda/).*(?=-AZInfo) regex to get the
        # stack name from the log group name check if this portion is in
        # the stack_dict dictionary if so delete the corresponding
        # log Group

        # Log Dict is a dictionary, with one key: logGroups which returns a
        # list of dictionaries that contains info about the log
        # groups log_dict['logGroups'] for a given element (in the list)
        # is a dictionary with the following keys:
        # metricFilterCount, creationtime, logGroupName, arn

        for x in log_dict['logGroups']:
            log_group_name = x['logGroupName']
            # Check if Nite is in the name if so delete the log group
            # otherwise check if the stacks status is delete complete
            # if so delete the log finally check if the stack is
            # not listed which means that the stack was deleted
            # more than 90 days ago so delete the log

            nite = re.findall('Nite', log_group_name)
            if len(nite) == 1 and nite[0] == "Nite":
                connection_logs.delete_log_group(log_group_name)
                print('Deleting: {}\n'.format(log_group_name))
            else:
                name = re.findall('(?<=aws/lambda/).*(?=-AZInfo)',
                                  log_group_name)

                stack_name = None
                if len(name) == 1:
                    stack_name = name[0]

                if (stack_name in stack_dict and
                   stack_dict[stack_name].stack_status == "DELETE_COMPLETE"):
                    connection_logs.delete_log_group(log_group_name)
                    print('Deleting: {}\n'.format(log_group_name))
                elif stack_name is not None and stack_name not in stack_dict:
                    connection_logs.delete_log_group(log_group_name)
                    print('deleting: {}\n'.format(log_group_name))
                else:
                    print('Not Deleting: {}\n'.format(log_group_name))

        next_token_log = log_dict.get('nextToken')
        log_dict = connection_logs.describe_log_groups(
                next_token=next_token_log)

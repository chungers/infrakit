
def parameter_groups():
    return [{"Label": {"default": "Swarm Size"},
             "Parameters": ["ManagerSize", "ClusterSize"]
             },
            {"Label": {"default": "Swarm Properties"},
             "Parameters": ["KeyName", "EnableSystemPrune", "EnableCloudWatchLogs", "EnableCloudStorEfs"]
             },
            {"Label": {"default": "Swarm Manager Properties"},
             "Parameters": ["ManagerInstanceType", "ManagerDiskSize", "ManagerDiskType"]
             },
            {"Label": {"default": "Swarm Worker Properties"},
             "Parameters": ["InstanceType", "WorkerDiskSize", "WorkerDiskType"]
             }]


def metadata(template, parameter_groups, parameter_labels):
    template.add_metadata({
        'AWS::CloudFormation::Interface': {
            'ParameterGroups': parameter_groups,
            'ParameterLabels': parameter_labels
        }
    })

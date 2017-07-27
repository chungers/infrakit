from troposphere import Parameter
from constants import ALL_INSTANCE_TYPES


def add_parameter_keyname(template):
    template.add_parameter(Parameter(
        "KeyName",
        Description="Name of an existing EC2 KeyPair to enable SSH "
                    "access to the instances",
        Type='AWS::EC2::KeyPair::KeyName',
        ConstraintDescription="Must be the name of an existing EC2 KeyPair"
    ))
    return ('KeyName', {"default": "Which SSH key to use?"})


def add_parameter_instancetype(template, default_instance_type=None,
                               instance_types=None):
    if not default_instance_type:
        default_instance_type = 't2.micro'
    if not instance_types:
        instance_types = ALL_INSTANCE_TYPES
    template.add_parameter(Parameter(
        'InstanceType',
        Type='String',
        Description='EC2 HVM instance type (t2.micro, m3.medium, etc).',
        Default=default_instance_type,
        AllowedValues=instance_types,
        ConstraintDescription='Must be a valid EC2 HVM instance type.',
    ))
    return ('InstanceType', {"default": "Agent worker instance type?"})


def add_parameter_manager_instancetype(template, default_instance_type=None,
                                       instance_types=None):
    if not default_instance_type:
        default_instance_type = 't2.micro'
    if not instance_types:
        instance_types = ALL_INSTANCE_TYPES

    template.add_parameter(Parameter(
        'ManagerInstanceType',
        Type='String',
        Description='EC2 HVM instance type (t2.micro, m3.medium, etc).',
        Default=default_instance_type,
        AllowedValues=instance_types,
        ConstraintDescription='Must be a valid EC2 HVM instance type.',
    ))
    return ('ManagerInstanceType', {"default": "Swarm manager instance type?"})


def add_parameter_cluster_size(template):
    template.add_parameter(Parameter(
        'ClusterSize',
        Type='Number',
        Default="5",
        MinValue="0",
        MaxValue="1000",
        Description="Number of worker nodes in the Swarm (0-1000)."))
    return ('ClusterSize', {"default": "Number of Swarm worker nodes?"})


def add_parameter_worker_disk_size(template):
    template.add_parameter(Parameter(
        'WorkerDiskSize',
        Type='Number',
        Default="20",
        MinValue="20",
        MaxValue="1024",
        Description="Size of Workers's ephemeral storage volume in GiB"))
    return ('WorkerDiskSize', {"default": "Worker ephemeral storage volume size?"})


def add_parameter_worker_disk_type(template):
    template.add_parameter(Parameter(
        'WorkerDiskType',
        Type='String',
        Default='standard',
        AllowedValues=["standard", "gp2"],
        Description="Worker ephemeral storage volume type"))
    return ('WorkerDiskType', {"default": "Worker ephemeral storage volume type"})


def add_parameter_manager_size(template, allowed_values=None):
    if not allowed_values:
        allowed_values = ["1", "3", "5"]
    allowed_value_str = ", ".join(allowed_values)
    template.add_parameter(Parameter(
        'ManagerSize',
        Type='Number',
        Default="3",
        AllowedValues=allowed_values,
        Description="Number of Swarm manager nodes ({})".format(
            allowed_value_str)))
    return ('ManagerSize', {"default": "Number of Swarm managers?"})


def add_parameter_manager_disk_size(template):
    template.add_parameter(Parameter(
        'ManagerDiskSize',
        Type='Number',
        Default="20",
        MinValue="20",
        MaxValue="1024",
        Description="Size of Manager's ephemeral storage volume in GiB"))
    return ('ManagerDiskSize', {"default": "Manager ephemeral storage volume size?"})


def add_parameter_manager_disk_type(template):
    template.add_parameter(Parameter(
        'ManagerDiskType',
        Type='String',
        Default='standard',
        AllowedValues=["standard", "gp2"],
        Description="Manager ephemeral storage volume type"))
    return ('ManagerDiskType', {"default": "Manager ephemeral storage volume type"})


def add_parameter_enable_system_prune(template):
    template.add_parameter(Parameter(
        'EnableSystemPrune',
        Type='String',
        Default='no',
        AllowedValues=["no", "yes"],
        Description="Cleans up unused images, containers, networks and volumes"))
    return ('EnableSystemPrune', {"default": "Enable daily resource cleanup?"})


def add_parameter_enable_cloudwatch_logs(template, default=None):
    if not default:
        default = 'yes'
    template.add_parameter(Parameter(
        'EnableCloudWatchLogs',
        Type='String',
        Default=default,
        AllowedValues=["no", "yes"],
        Description="Send all Container logs to CloudWatch"))
    return ('EnableCloudWatchLogs', {"default": "Use Cloudwatch for container logging?"})


def add_parameter_enable_cloudstor_efs(template):
    template.add_parameter(Parameter(
        'EnableCloudStorEfs',
        Type='String',
        Default='no',
        AllowedValues=["no", "yes"],
        Description="Create CloudStor EFS mount targets"))
    return ('EnableCloudStorEfs', {"default": "Create EFS prerequsities for CloudStor?"})


def add_parameter_enable_ebs_optimized(template, default=None):
    if not default:
        default = 'no'
    template.add_parameter(Parameter(
        'EnableEbsOptimized',
        Type='String',
        Default=default,
        AllowedValues=["no", "yes"],
        Description="Specifies whether the launch configuration is optimized for EBS I/O"))
    return ('EnableEbsOptimized', {"default": "Enable EBS I/O optimization?"})

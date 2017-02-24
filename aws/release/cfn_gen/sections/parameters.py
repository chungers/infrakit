from troposphere import Parameter


def add_parameter_keyname(template):
    template.add_parameter(Parameter(
        "KeyName",
        Description="Name of an existing EC2 KeyPair to enable SSH "
                    "access to the instances",
        Type='AWS::EC2::KeyPair::KeyName',
        ConstraintDescription="Must be the name of an existing EC2 KeyPair"
    ))
    return ('KeyName', {"default": "Which SSH key to use?"})


def add_parameter_instancetype(template):
    template.add_parameter(Parameter(
        'InstanceType',
        Type='String',
        Description='EC2 HVM instance type (t2.micro, m3.medium, etc).',
        Default='t2.micro',
        AllowedValues=[
            "t2.micro", "t2.small", "t2.medium", "t2.large", "t2.xlarge", "t2.2xlarge",
            "m4.large", "m4.xlarge", "m4.2xlarge", "m4.4xlarge", "m4.10xlarge", "m3.medium",
            "m3.large", "m3.xlarge", "m3.2xlarge", "c4.large", "c4.xlarge", "c4.2xlarge",
            "c4.4xlarge", "c4.8xlarge", "c3.large", "c3.xlarge", "c3.2xlarge", "c3.4xlarge",
            "c3.8xlarge", "r3.large", "r3.xlarge", "r3.2xlarge", "r3.4xlarge", "r3.8xlarge",
            "r4.large", "r4.xlarge", "r4.2xlarge", "r4.4xlarge", "r4.8xlarge", "r4.16xlarge",
            "i2.xlarge", "i2.2xlarge", "i2.4xlarge", "i2.8xlarge"
        ],
        ConstraintDescription='Must be a valid EC2 HVM instance type.',
    ))
    return ('InstanceType', {"default": "Agent worker instance type?"})


def add_parameter_manager_instancetype(template):
    template.add_parameter(Parameter(
        'ManagerInstanceType',
        Type='String',
        Description='EC2 HVM instance type (t2.micro, m3.medium, etc).',
        Default='t2.micro',
        AllowedValues=[
            "t2.micro", "t2.small", "t2.medium", "t2.large", "t2.xlarge", "t2.2xlarge",
            "m4.large", "m4.xlarge", "m4.2xlarge", "m4.4xlarge", "m4.10xlarge", "m3.medium",
            "m3.large", "m3.xlarge", "m3.2xlarge", "c4.large", "c4.xlarge", "c4.2xlarge",
            "c4.4xlarge", "c4.8xlarge", "c3.large", "c3.xlarge", "c3.2xlarge", "c3.4xlarge",
            "c3.8xlarge", "r3.large", "r3.xlarge", "r3.2xlarge", "r3.4xlarge", "r3.8xlarge",
            "r4.large", "r4.xlarge", "r4.2xlarge", "r4.4xlarge", "r4.8xlarge", "r4.16xlarge",
            "i2.xlarge", "i2.2xlarge", "i2.4xlarge", "i2.8xlarge"
        ],
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


def add_parameter_manager_size(template):
    template.add_parameter(Parameter(
        'ManagerSize',
        Type='Number',
        Default="3",
        AllowedValues=["1", "3", "5"],
        Description="Number of Swarm manager nodes (1, 3, 5)"))
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


def add_parameter_enable_cloudwatch_logs(template):
    template.add_parameter(Parameter(
        'EnableCloudWatchLogs',
        Type='String',
        Default='yes',
        AllowedValues=["no", "yes"],
        Description="Send all Container logs to CloudWatch"))
    return ('EnableCloudWatchLogs', {"default": "Use Cloudwatch for container logging?"})

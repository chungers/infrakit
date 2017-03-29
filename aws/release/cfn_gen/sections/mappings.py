
def add_mapping_vpc_cidrs(template):
    template.add_mapping("VpcCidrs", {
        "vpc": {
            "cidr": "172.31.0.0/16"
        },
        "pubsubnet1": {
            "cidr": "172.31.0.0/20"
        },
        "pubsubnet2": {
            "cidr": "172.31.16.0/20"
        },
        "pubsubnet3": {
            "cidr": "172.31.32.0/20"
        },
        "pubsubnet4": {
            "cidr": "172.31.48.0/20"
        }
    })


def add_mapping_amis(template, amis):
    template.add_mapping("AWSRegionArch2AMI", amis)


def add_mapping_instance_type_2_arch(template):
    template.add_mapping("AWSInstanceType2Arch", {
        "t2.micro": {
            "Arch": "HVM64"
        },
        "t2.small": {
            "Arch": "HVM64"
        },
        "t2.medium": {
            "Arch": "HVM64"
        },
        "t2.large": {
            "Arch": "HVM64"
        },
        "t2.xlarge": {
            "Arch": "HVM64"
        },
        "t2.2xlarge": {
            "Arch": "HVM64"
        },
        "m3.medium": {
            "Arch": "HVM64"
        },
        "m3.large": {
            "Arch": "HVM64"
        },
        "m3.xlarge": {
            "Arch": "HVM64"
        },
        "m3.2xlarge": {
            "Arch": "HVM64"
        },
        "m4.large": {
            "Arch": "HVM64"
        },
        "m4.xlarge": {
            "Arch": "HVM64"
        },
        "m4.2xlarge": {
            "Arch": "HVM64"
        },
        "m4.4xlarge": {
            "Arch": "HVM64"
        },
        "m4.10xlarge": {
            "Arch": "HVM64"
        },
        "c3.large": {
            "Arch": "HVM64"
        },
        "c3.xlarge": {
            "Arch": "HVM64"
        },
        "c3.2xlarge": {
            "Arch": "HVM64"
        },
        "c3.4xlarge": {
            "Arch": "HVM64"
        },
        "c3.8xlarge": {
            "Arch": "HVM64"
        },
        "c4.large": {
            "Arch": "HVM64"
        },
        "c4.xlarge": {
            "Arch": "HVM64"
        },
        "c4.2xlarge": {
            "Arch": "HVM64"
        },
        "c4.4xlarge": {
            "Arch": "HVM64"
        },
        "c4.8xlarge": {
            "Arch": "HVM64"
        },
        "g2.2xlarge": {
            "Arch": "HVMG2"
        },
        "r3.large": {
            "Arch": "HVM64"
        },
        "r3.xlarge": {
            "Arch": "HVM64"
        },
        "r3.2xlarge": {
            "Arch": "HVM64"
        },
        "r3.4xlarge": {
            "Arch": "HVM64"
        },
        "r3.8xlarge": {
            "Arch": "HVM64"
        },
        "r4.large": {
            "Arch": "HVM64"
        },
        "r4.xlarge": {
            "Arch": "HVM64"
        },
        "r4.2xlarge": {
            "Arch": "HVM64"
        },
        "r4.4xlarge": {
            "Arch": "HVM64"
        },
        "r4.8xlarge": {
            "Arch": "HVM64"
        },
        "r4.16xlarge": {
            "Arch": "HVM64"
        },
        "i2.xlarge": {
            "Arch": "HVM64"
        },
        "i2.2xlarge": {
            "Arch": "HVM64"
        },
        "i2.4xlarge": {
            "Arch": "HVM64"
        },
        "i2.8xlarge": {
            "Arch": "HVM64"
        },
        "i3.large": {
            "Arch": "HVM64"
        },
        "i3.xlarge": {
            "Arch": "HVM64"
        },
        "i3.2xlarge": {
            "Arch": "HVM64"
        },
        "i3.4xlarge": {
            "Arch": "HVM64"
        },
        "i3.8xlarge": {
            "Arch": "HVM64"
        },
        "i3.16xlarge": {
            "Arch": "HVM64"
        },
        "d2.xlarge": {
            "Arch": "HVM64"
        },
        "d2.2xlarge": {
            "Arch": "HVM64"
        },
        "d2.4xlarge": {
            "Arch": "HVM64"
        },
        "d2.8xlarge": {
            "Arch": "HVM64"
        },
        "hi1.4xlarge": {
            "Arch": "HVM64"
        },
        "hs1.8xlarge": {
            "Arch": "HVM64"
        },
        "cr1.8xlarge": {
            "Arch": "HVM64"
        },
        "cc2.8xlarge": {
            "Arch": "HVM64"
        }
    })


def add_mapping_version(template, docker_version,
                        d4a_version, d4a_addon, channel, has_ddc,
                        extra_data=None):
    """ Use extra_data to pass in anythig extra you want to add
    to the version dict """
    if has_ddc:
        ddc = 'yes'
    else:
        ddc = 'no'
    data = {
        "docker": docker_version,
        "forAws": d4a_version,
        "channel": channel,
        "addOn": d4a_addon,
        'HasDDC': ddc,
    }
    if extra_data:
        data.update(extra_data)

    template.add_mapping("DockerForAWS", {
        "version": data
    })


def aws2az_data():
    return {
        "ap-northeast-1": {
            "Name": "Tokyo",
            "EFSSupport": "no",
            "LambdaSupport": "yes",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "ap-northeast-2": {
            "Name": "Seoul",
            "EFSSupport": "no",
            "LambdaSupport": "yes",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "ap-south-1": {
            "Name": "Mumbai",
            "EFSSupport": "no",
            "LambdaSupport": "yes",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "ap-southeast-1": {
            "Name": "Singapore",
            "EFSSupport": "no",
            "LambdaSupport": "yes",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "ap-southeast-2": {
            "Name": "Sydney",
            "EFSSupport": "yes",
            "LambdaSupport": "yes",
            "NumAZs": "3",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "2"
        },
        "ca-central-1": {
            "Name": "Central",
            "EFSSupport": "no",
            "LambdaSupport": "no",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "eu-central-1": {
            "Name": "Frankfurt",
            "EFSSupport": "no",
            "LambdaSupport": "yes",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "eu-west-1": {
            "Name": "Ireland",
            "EFSSupport": "yes",
            "LambdaSupport": "yes",
            "NumAZs": "3",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "2"
        },
        "eu-west-2": {
            "Name": "London",
            "EFSSupport": "no",
            "LambdaSupport": "yes",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "sa-east-1": {
            "Name": "Sao Paulo",
            "EFSSupport": "no",
            "LambdaSupport": "no",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "us-east-1": {
            "Name": "N. Virgina",
            "EFSSupport": "yes",
            "LambdaSupport": "yes",
            "NumAZs": "4",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "2"
        },
        "us-east-2": {
            "Name": "Ohio",
            "EFSSupport": "yes",
            "LambdaSupport": "yes",
            "NumAZs": "3",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "2"
        },
        "us-west-1": {
            "Name": "N. California",
            "EFSSupport": "no",
            "LambdaSupport": "yes",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        },
        "us-west-2": {
            "Name": "Oregon",
            "EFSSupport": "yes",
            "LambdaSupport": "yes",
            "NumAZs": "3",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "2"
        },
        "us-gov-west-1": {
            "Name": "GovCloud",
            "EFSSupport": "no",
            "LambdaSupport": "no",
            "NumAZs": "2",
            "AZ0": "0",
            "AZ1": "1",
            "AZ2": "0"
        }
    }

# DTR specific AWS elements

from troposphere import FindInMap, GetAtt, Ref, If, Join, Parameter
from troposphere.autoscaling import LifecycleHook, AutoScalingGroup, LaunchConfiguration
from troposphere.policies import (
    AutoScalingRollingUpdate, UpdatePolicy, CreationPolicy, ResourceSignal)
from troposphere.ec2 import BlockDeviceMapping, EBSBlockDevice
from troposphere.elasticloadbalancing import (
    LoadBalancer, HealthCheck, ConnectionSettings, Listener)
from troposphere.ec2 import SecurityGroup, SecurityGroupRule
from troposphere.iam import Role, InstanceProfile, PolicyType
from troposphere.iam import Policy as TroposherePolicy

from awacs.aws import Allow, Statement, Principal, Policy
from awacs.sts import AssumeRole


def dtr_node_upgrade_hook(template):
    """
        The DTR worker node upgrade hook
    """
    template.add_resource(LifecycleHook(
        "DTRWorkerUpgradeHook",
        DependsOn="SwarmSQS",
        AutoScalingGroupName=Ref("DTRAsg"),
        LifecycleTransition="autoscaling:EC2_INSTANCE_TERMINATING",
        NotificationTargetARN=GetAtt("SwarmSQS", "Arn"),
        RoleARN=GetAtt("ProxyRole", "Arn")
    ))


def dtr_autoscalegroup(template, launch_config_name):
    """
        DTR autoscaping group
    """
    template.add_resource(AutoScalingGroup(
        "DTRAsg",
        DependsOn="ManagerAsg",
        Tags=[
            {'Key': "Name",
             'Value': Join("-", [Ref("AWS::StackName"), "dtr"]),
             'PropagateAtLaunch': True},
            {'Key': "swarm-node-type",
             'Value': "worker",
             'PropagateAtLaunch': True},
            {'Key': "swarm-stack-id",
             'Value': Ref("AWS::StackId"),
             'PropagateAtLaunch': True},
            {'Key': "DOCKER_FOR_AWS_VERSION",
             'Value': FindInMap("DockerForAWS", "version", "forAws"),
             'PropagateAtLaunch': True},
            {'Key': "DOCKER_VERSION",
             'Value': FindInMap("DockerForAWS", "version", "docker"),
             'PropagateAtLaunch': True}
        ],
        LaunchConfigurationName=Ref(launch_config_name),
        MinSize=0,
        MaxSize=3,
        DesiredCapacity=3,
        VPCZoneIdentifier=[
            If("HasOnly2AZs",
               Join(",", [Ref("PubSubnetAz1"), Ref("PubSubnetAz2")]),
               Join(",", [Ref("PubSubnetAz1"), Ref("PubSubnetAz2"),
                    Ref("PubSubnetAz3")]))
        ],
        LoadBalancerNames=[Ref("DTRLoadBalancer")],
        HealthCheckType="ELB",
        HealthCheckGracePeriod=1200,
        UpdatePolicy=UpdatePolicy(
            AutoScalingRollingUpdate=AutoScalingRollingUpdate(
                PauseTime='PT1H',
                MinInstancesInService=2,
                MaxBatchSize='1',
                WaitOnResourceSignals=True
            )
        ),
        CreationPolicy=CreationPolicy(
            ResourceSignal=ResourceSignal(
                Timeout='PT2H',
                Count=3
                )
        )
    ))


def dtr_asg_launch_config(template, user_data,
                          launch_config_name="DTRLaunchConfig"):
    """
        DTR ASG launch config
    """
    template.add_resource(LaunchConfiguration(
        launch_config_name,
        DependsOn="ManagerAsg",
        UserData=user_data,  # compile_worker_node_userdata(),
        ImageId=FindInMap(
            "AWSRegionArch2AMI",
            Ref("AWS::Region"),
            FindInMap("AWSInstanceType2Arch", Ref("InstanceType"), "Arch")
        ),
        KeyName=Ref("KeyName"),
        BlockDeviceMappings=[
            BlockDeviceMapping(
                DeviceName="/dev/xvdb",
                Ebs=EBSBlockDevice(
                    VolumeSize=Ref("DTRDiskSize"),
                    VolumeType=Ref("DTRDiskType")
                )
            ),
        ],
        SecurityGroups=[Ref("NodeVpcSG")],
        InstanceType=Ref("ManagerInstanceType"),
        AssociatePublicIpAddress=True,
        IamInstanceProfile=Ref("DTRInstanceProfile"),
    ))


def add_resource_ddc_dtr_lb(template, create_vpc, extra_listeners=None):
    """
        Add the DTR LB
    """
    if create_vpc:
        depends = ["AttachGateway", "DTRLoadBalancerSG",
                   "PubSubnetAz1", "PubSubnetAz2", "PubSubnetAz3"]
    else:
        depends = ["DTRLoadBalancerSG"]

    listener_list = []
    listener_list.append(Listener(
        LoadBalancerPort="443",
        InstancePort="12391",
        Protocol="TCP"
    ),)
    if extra_listeners:
        listener_list.extend(extra_listeners)

    template.add_resource(LoadBalancer(
        "DTRLoadBalancer",
        DependsOn=depends,
        ConnectionSettings=ConnectionSettings(IdleTimeout=1800),
        Subnets=If("HasOnly2AZs",
                   [Ref("PubSubnetAz1"), Ref("PubSubnetAz2")],
                   [Ref("PubSubnetAz1"), Ref("PubSubnetAz2"),
                    Ref("PubSubnetAz3")]),
        HealthCheck=HealthCheck(
            Target="HTTPS:12391/health",
            HealthyThreshold="2",
            UnhealthyThreshold="10",
            Interval="300",
            Timeout="10",
        ),
        Listeners=listener_list,
        CrossZone=True,
        SecurityGroups=[Ref("DTRLoadBalancerSG")],
        Tags=[
            {'Key': "Name",
             'Value': Join("-", [Ref("AWS::StackName"), "ELB-DTR"])}
        ]
    ))


def add_resource_ddc_dtr_lb_sg(template, create_vpc):
    """
        DTR LB security group
    """
    sg = SecurityGroup(
        "DTRLoadBalancerSG",
        VpcId=Ref("Vpc"),
        GroupDescription="DTR Load Balancer SecurityGroup",
        SecurityGroupIngress=[SecurityGroupRule(
            IpProtocol='tcp',
            FromPort='443',
            ToPort='443',
            CidrIp="0.0.0.0/0",
        )]
    )
    # have to do this, because DependsOn can't be None or ""
    if create_vpc:
        sg.DependsOn = "Vpc"
    template.add_resource(sg)


def dtr_disk_type(template):
    template.add_parameter(Parameter(
        'DTRDiskType',
        Type='String',
        Default='standard',
        AllowedValues=["standard", "gp2"],
        Description="DTR ephemeral storage volume type"))
    return ('DTRDiskType', {"default": "DTR ephemeral storage volume type"})


def dtr_disk_size(template):
    template.add_parameter(Parameter(
        'DTRDiskSize',
        Type='Number',
        Default="20",
        MinValue="20",
        MaxValue="1024",
        Description="Size of DTR's ephemeral storage volume in GiB"))
    return ('DTRDiskSize', {"default": "DTR ephemeral storage volume size?"})


def dtr_iam_role(template):
    """
    DTR IAM Role for DTR nodes
    """
    template.add_resource(Role(
        "DTRRole",
        AssumeRolePolicyDocument=Policy(
            Version="2012-10-17",
            Statement=[
                Statement(
                    Effect=Allow,
                    Action=[AssumeRole],
                    Principal=Principal(
                        "Service", ["ec2.amazonaws.com",
                                    "autoscaling.amazonaws.com"])
                )
            ]
        ),
        Path="/"
    ))


def iam_dtr_instance_profile(template):
    """
        DTR IAM instance profiles
    """
    template.add_resource(InstanceProfile(
        "DTRInstanceProfile",
        DependsOn="DTRRole",
        Path="/",
        Roles=[Ref("DTRRole")],
    ))


def add_resource_s3_ddc_bucket_policy(template):
    """
        S3 IAM POLICY
    """
    template.add_resource(PolicyType(
        "S3Policies",
        DependsOn="DTRRole",
        PolicyName="S3-DDC-Policy",
        Roles=[Ref("DTRRole")],
        PolicyDocument={
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Action": [
                    "s3:ListBucket",
                    "s3:GetBucketLocation",
                    "s3:ListBucketMultipartUploads"
                ],
                "Resource": Join(
                    "", ["arn:aws:s3:::", Ref("DDCBucket")])
            }, {
                "Effect": "Allow",
                "Action": [
                    "s3:PutObject",
                    "s3:GetObject",
                    "s3:DeleteObject",
                    "s3:ListMultipartUploadParts",
                    "s3:AbortMultipartUpload"
                ],
                "Resource": Join(
                    "", ["arn:aws:s3:::", Ref("DDCBucket"), "/*"])
            }
            ],
        }
    ))

from troposphere import GetAtt, Ref, Join
from troposphere.iam import Role, InstanceProfile, PolicyType

from awacs.aws import Allow, Statement, Principal, Policy
from awacs.sts import AssumeRole


def add_resource_proxy_role(template):
    """
    "ProxyRole": {
        "Type": "AWS::IAM::Role",
        "Properties": {
        "AssumeRolePolicyDocument": {
            "Version" : "2012-10-17",
            "Statement": [ {
            "Effect": "Allow",
            "Principal": {
                "Service": [ "ec2.amazonaws.com", "autoscaling.amazonaws.com" ]
            },
            "Action": [ "sts:AssumeRole" ]
            } ]
        },
        "Path": "/"
            }
    },
    """
    template.add_resource(Role(
        "ProxyRole",
        AssumeRolePolicyDocument=Policy(
            Version="2012-10-17",
            Statement=[
                Statement(
                    Effect=Allow,
                    Action=[AssumeRole],
                    Principal=Principal(
                        "Service", ["ec2.amazonaws.com", "autoscaling.amazonaws.com"])
                )
            ]
        ),
        Path="/"
    ))


def add_resource_IAM_dyn_policy(template):
    """
    "DynDBPolicies": {
        "DependsOn": ["ProxyRole", "SwarmDynDBTable"],
        "Type": "AWS::IAM::Policy",
        "Properties": {
        "PolicyName": "dyndb-getput",
        "PolicyDocument": {
            "Version" : "2012-10-17",
            "Statement": [ {
                "Effect": "Allow",
                "Action": [
                    "dynamodb:PutItem",
                    "dynamodb:DeleteItem",
                    "dynamodb:GetItem",
                    "dynamodb:UpdateItem",
                    "dynamodb:Query"
                ],
                "Resource": { "Fn::Join": ["", ["arn:aws:dynamodb:",
                    { "Ref": "AWS::Region" }, ":", { "Ref": "AWS::AccountId" }, ":table/",
                    { "Ref": "SwarmDynDBTable" }]] }
            } ]
        },
        "Roles": [ {
            "Ref": "ProxyRole"
        } ]
            }
    },
    """
    template.add_resource(PolicyType(
        "DynDBPolicies",
        DependsOn=["ProxyRole", "SwarmDynDBTable"],
        PolicyName="dyndb-getput",
        Roles=[Ref("ProxyRole")],
        PolicyDocument={
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Action": [
                    "dynamodb:PutItem",
                    "dynamodb:DeleteItem",
                    "dynamodb:GetItem",
                    "dynamodb:UpdateItem",
                    "dynamodb:Query"
                ],
                "Resource": Join(
                    "", ["arn:aws:dynamodb:",
                         Ref("AWS::Region"), ":",
                         Ref("AWS::AccountId"), ":table/",
                         Ref("SwarmDynDBTable")])
            }],
        }
    ))


def add_resource_iam_swarm_api_policy(template):
    """
    "SwarmAPIPolicy": {
        "DependsOn": "ProxyRole",
        "Type": "AWS::IAM::Policy",
        "Properties": {
        "PolicyName": "swarm-policy",
        "PolicyDocument": {
            "Version" : "2012-10-17",
            "Statement": [ {
                "Effect": "Allow",
                "Action": [
                    "ec2:DescribeInstances",
                    "ec2:DescribeVpcAttribute",
                    "logs:CreateLogStream",
                    "logs:PutLogEvents"
                ],
                "Resource": "*"
            } ]
        },
        "Roles": [ {
            "Ref": "ProxyRole"
        } ]
            }
    },
    """
    template.add_resource(PolicyType(
        "SwarmAPIPolicy",
        DependsOn="ProxyRole",
        PolicyName="swarm-policy",
        Roles=[Ref("ProxyRole")],
        PolicyDocument={
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Action": [
                    "ec2:DescribeInstances",
                    "ec2:DescribeVpcAttribute",
                    "logs:CreateLogStream",
                    "logs:PutLogEvents"
                ],
                "Resource": "*"
            }],
        }
    ))


def add_resource_iam_sqs_policy(template):
    """
    "SwarmSQSPolicy": {
        "DependsOn": ["ProxyRole", "SwarmSQS"],
        "Type": "AWS::IAM::Policy",
        "Properties": {
        "PolicyName": "swarm-sqs-policy",
        "PolicyDocument": {
            "Version" : "2012-10-17",
            "Statement": [ {
                "Effect": "Allow",
                "Action": "sqs:*",
                "Resource": { "Fn::GetAtt" : ["SwarmSQS", "Arn"]}
            } ]
        },
        "Roles": [ {
            "Ref": "ProxyRole"
        } ]
            }
    },
    """
    template.add_resource(PolicyType(
        "SwarmSQSPolicy",
        DependsOn=["ProxyRole", "SwarmSQS"],
        PolicyName="swarm-sqs-policy",
        Roles=[Ref("ProxyRole")],
        PolicyDocument={
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Action": "sqs:*",
                "Resource": GetAtt("SwarmSQS", "Arn")
            }],
        }
    ))


def add_resource_iam_sqs_cleanup_policy(template):
    """
    "SwarmSQSCleanupPolicy": {
        "DependsOn": ["ProxyRole", "SwarmSQSCleanup"],
        "Type": "AWS::IAM::Policy",
        "Properties": {
        "PolicyName": "swarm-sqs-cleanup-policy",
        "PolicyDocument": {
            "Version" : "2012-10-17",
            "Statement": [ {
                "Effect": "Allow",
                "Action": "sqs:*",
                "Resource": { "Fn::GetAtt" : ["SwarmSQSCleanup", "Arn"]}
            } ]
        },
        "Roles": [ {
            "Ref": "ProxyRole"
        } ]
            }
    },
    """
    template.add_resource(PolicyType(
        "SwarmSQSCleanupPolicy",
        DependsOn=["ProxyRole", "SwarmSQSCleanup"],
        PolicyName="swarm-sqs-cleanup-policy",
        Roles=[Ref("ProxyRole")],
        PolicyDocument={
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Action": "sqs:*",
                "Resource": GetAtt("SwarmSQSCleanup", "Arn")
            }],
        }
    ))


def add_resource_iam_autoscale_policy(template):
    """
    "SwarmAutoscalePolicy": {
        "Type": "AWS::IAM::Policy",
        "Properties": {
        "PolicyName": "swarm-autoscale-policy",
        "PolicyDocument": {
            "Version" : "2012-10-17",
            "Statement": [ {
                "Effect": "Allow",
                "Action": "autoscaling:*",
                "Resource": "*"
            } ]
        },
        "Roles": [ {
            "Ref": "ProxyRole"
        } ]
            }
    },
    """
    template.add_resource(PolicyType(
        "SwarmAutoscalePolicy",
        DependsOn="ProxyRole",
        PolicyName="swarm-autoscale-policy",
        Roles=[Ref("ProxyRole")],
        PolicyDocument={
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Action": "autoscaling:*",
                "Resource": "*"
            }],
        }
    ))


def add_resource_iam_elb_policy(template):
    """
    "ProxyPolicies": {
        "Type": "AWS::IAM::Policy",
        "Properties": {
        "PolicyName": "elb-update",
        "PolicyDocument": {
            "Version" : "2012-10-17",
            "Statement": [ {
            "Effect": "Allow",
            "Action": "elasticloadbalancing:*",
            "Resource": "*"
            } ]
        },
        "Roles": [ {
            "Ref": "ProxyRole"
        } ]
            }
    },
    """
    template.add_resource(PolicyType(
        "ProxyPolicies",
        DependsOn="ProxyRole",
        PolicyName="elb-update",
        Roles=[Ref("ProxyRole")],
        PolicyDocument={
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Action": "elasticloadbalancing:*",
                "Resource": "*"
            }],
        }
    ))


def add_resource_iam_instance_profile(template):
    """
    "ProxyInstanceProfile": {
        "Type": "AWS::IAM::InstanceProfile",
        "Properties": {
        "Path": "/",
        "Roles": [ {
            "Ref": "ProxyRole"
        } ]
            }
    }
    """
    template.add_resource(InstanceProfile(
        "ProxyInstanceProfile",
        DependsOn="ProxyRole",
        Path="/",
        Roles=[Ref("ProxyRole")],
    ))

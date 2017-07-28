from troposphere import FindInMap, GetAtt, Ref

from troposphere.ec2 import SecurityGroup, SecurityGroupRule


def add_resource_swarm_wide_security_group(template, create_vpc):
    """
    "SwarmWideSG": {
        "DependsOn": "Vpc",
        "Type": "AWS::EC2::SecurityGroup",
        "Properties": {
            "VpcId": {
                "Ref": "Vpc"
            },
            "GroupDescription": "Swarm wide access",
            "SecurityGroupIngress": [
                {
                    "IpProtocol": "-1",
                    "FromPort": "0",
                    "ToPort": "65535",
                    "CidrIp": { "Fn::FindInMap" : [ "VpcCidrs", "vpc", "cidr" ] }
                }
            ]
        }
    },
    """
    if create_vpc:
        cidr = FindInMap("VpcCidrs", "vpc", "cidr")
    else:
        cidr = Ref("VpcCidr")

    sg = SecurityGroup(
        "SwarmWideSG",
        VpcId=Ref("Vpc"),
        GroupDescription="Swarm wide access",
        SecurityGroupIngress=[SecurityGroupRule(
            IpProtocol='-1',
            FromPort='0',
            ToPort='65535',
            CidrIp=cidr,
        )]
    )
    # have to do this, because DependsOn can't be None or ""
    if create_vpc:
        sg.DependsOn = "Vpc"
    template.add_resource(sg)


def add_resource_external_lb_sg(template, create_vpc):
    """
    "ExternalLoadBalancerSG": {
        "DependsOn": "Vpc",
        "Type": "AWS::EC2::SecurityGroup",
        "Properties": {
            "VpcId": {
                "Ref": "Vpc"
            },
            "GroupDescription": "External Load Balancer SecurityGroup",
            "SecurityGroupIngress": [
                {"IpProtocol": "-1","FromPort": "0","ToPort": "65535","CidrIp": "0.0.0.0/0"}
            ]
        }
    },
    """

    sg = SecurityGroup(
        "ExternalLoadBalancerSG",
        VpcId=Ref("Vpc"),
        GroupDescription="External Load Balancer SecurityGroup",
        SecurityGroupIngress=[SecurityGroupRule(
            IpProtocol='-1',
            FromPort='0',
            ToPort='65535',
            CidrIp="0.0.0.0/0",
        )]
    )
    # have to do this, because DependsOn can't be None or ""
    if create_vpc:
        sg.DependsOn = "Vpc"
    template.add_resource(sg)


def add_resource_manager_security_group(template, use_ssh_cidr=False):
    """
    "ManagerVpcSG": {
        "DependsOn": "NodeVpcSG",
        "Type": "AWS::EC2::SecurityGroup",
        "Properties": {
            "VpcId": {
                "Ref": "Vpc"
            },
            "GroupDescription": "Manager SecurityGroup",
            "SecurityGroupIngress": [
                {"IpProtocol": "tcp", "FromPort": "22","ToPort": "22","CidrIp": "0.0.0.0/0"},
                {"IpProtocol" : "tcp", "FromPort" : "2377", "ToPort" : "2377",
                    "SourceSecurityGroupId" : { "Fn::GetAtt" : [ "NodeVpcSG", "GroupId" ] } },
                {"IpProtocol" : "udp", "FromPort" : "4789", "ToPort" : "4789",
                    "SourceSecurityGroupId" : { "Fn::GetAtt" : [ "NodeVpcSG", "GroupId" ] } },
                {"IpProtocol" : "tcp", "FromPort" : "7946", "ToPort" : "7946",
                    "SourceSecurityGroupId" : { "Fn::GetAtt" : [ "NodeVpcSG", "GroupId" ] } },
                {"IpProtocol" : "udp", "FromPort" : "7946", "ToPort" : "7946",
                    "SourceSecurityGroupId" : { "Fn::GetAtt" : [ "NodeVpcSG", "GroupId" ] } }
            ]
        }
    },
    """
    if use_ssh_cidr:
        ssh_cidr = Ref("RemoteSSH")
    else:
        ssh_cidr = "0.0.0.0/0"
    template.add_resource(SecurityGroup(
        "ManagerVpcSG",
        DependsOn="NodeVpcSG",
        VpcId=Ref("Vpc"),
        GroupDescription="Manager SecurityGroup",
        SecurityGroupIngress=[
            SecurityGroupRule(
                IpProtocol='tcp',
                FromPort='22',
                ToPort='22',
                CidrIp=ssh_cidr),
            SecurityGroupRule(
                IpProtocol='50',
                SourceSecurityGroupId=GetAtt("NodeVpcSG", "GroupId")),
            SecurityGroupRule(
                IpProtocol='tcp',
                FromPort='2377',
                ToPort='2377',
                SourceSecurityGroupId=GetAtt("NodeVpcSG", "GroupId")),
            SecurityGroupRule(
                IpProtocol='udp',
                FromPort='4789',
                ToPort='4789',
                SourceSecurityGroupId=GetAtt("NodeVpcSG", "GroupId")),
            SecurityGroupRule(
                IpProtocol='tcp',
                FromPort='7946',
                ToPort='7946',
                SourceSecurityGroupId=GetAtt("NodeVpcSG", "GroupId")),
            SecurityGroupRule(
                IpProtocol='udp',
                FromPort='7946',
                ToPort='7946',
                SourceSecurityGroupId=GetAtt("NodeVpcSG", "GroupId")),
        ],
        SecurityGroupEgress=[
            SecurityGroupRule(
                IpProtocol='50',
                SourceSecurityGroupId=GetAtt("NodeVpcSG", "GroupId")),
        ]
    ))


def add_resource_worker_security_group(template, create_vpc):
    """
    "NodeVpcSG": {
        "DependsOn": "Vpc",
        "Type": "AWS::EC2::SecurityGroup",
        "Properties": {
            "VpcId": {
                "Ref": "Vpc"
            },
            "GroupDescription": "Node SecurityGroup",
            "SecurityGroupIngress": [
                {
                    "IpProtocol": "-1",
                    "FromPort": "0",
                    "ToPort": "65535",
                    "CidrIp": { "Fn::FindInMap" : [ "VpcCidrs", "vpc", "cidr" ] }
                }
            ],
            "SecurityGroupEgress": [
                {"IpProtocol" : "icmp", "FromPort" : "8", "ToPort" : "0",
                    "CidrIp": "0.0.0.0/0" },
                {"IpProtocol" : "udp", "FromPort" : "0", "ToPort" : "65535",
                    "CidrIp": "0.0.0.0/0" },
                {"IpProtocol" : "tcp", "FromPort" : "0", "ToPort" : "2374",
                    "CidrIp": "0.0.0.0/0" },
                {"IpProtocol" : "tcp", "FromPort" : "2376", "ToPort" : "65535",
                    "CidrIp": "0.0.0.0/0" }
            ]
        }
    }
    """

    if create_vpc:
        cidr = FindInMap("VpcCidrs", "vpc", "cidr")
    else:
        cidr = Ref("VpcCidr")

    sg = SecurityGroup(
        "NodeVpcSG",
        VpcId=Ref("Vpc"),
        GroupDescription="Node SecurityGroup",
        SecurityGroupIngress=[
            SecurityGroupRule(
                IpProtocol='-1',
                FromPort='0',
                ToPort='65535',
                CidrIp=cidr),
            SecurityGroupRule(
                IpProtocol='50',
                CidrIp=cidr),
        ],
        SecurityGroupEgress=[
            SecurityGroupRule(
                IpProtocol='icmp',
                FromPort='8',
                ToPort='0',
                CidrIp="0.0.0.0/0"),
            SecurityGroupRule(
                IpProtocol='50',
                CidrIp="0.0.0.0/0"),
            SecurityGroupRule(
                IpProtocol='udp',
                FromPort='0',
                ToPort='65535',
                CidrIp="0.0.0.0/0"),
            SecurityGroupRule(
                IpProtocol='tcp',
                FromPort='0',
                ToPort='2374',
                CidrIp="0.0.0.0/0"),
            SecurityGroupRule(
                IpProtocol='tcp',
                FromPort='2376',
                ToPort='65535',
                CidrIp="0.0.0.0/0"),
        ]
    )
    # have to do this, because DependsOn can't be None or ""
    if create_vpc:
        sg.DependsOn = "Vpc"
    template.add_resource(sg)


def add_resource_ddc_ucp_lb_sg(template, create_vpc):
    """
    "UCPLoadBalancerSG": {
            "DependsOn": "Vpc",
            "Properties": {
                "GroupDescription": "UCP Load Balancer SecurityGroup",
                "SecurityGroupIngress": [
                    {
                        "CidrIp": "0.0.0.0/0",
                        "FromPort": "443",
                        "IpProtocol": "tcp",
                        "ToPort": "443"
                    }
                ],
                "VpcId": {
                    "Ref": "Vpc"
                }
            },
            "Type": "AWS::EC2::SecurityGroup"
        },
    """

    sg = SecurityGroup(
        "UCPLoadBalancerSG",
        VpcId=Ref("Vpc"),
        GroupDescription="UCP Load Balancer SecurityGroup",
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

from troposphere import FindInMap, Ref, Join, Select, GetAZs, If, GetAtt
from troposphere.ec2 import (
    VPC, Subnet, VPCGatewayAttachment, RouteTable, Route, SubnetRouteTableAssociation,
    InternetGateway)


def add_resource_vpc(template):
    template.add_resource(
        VPC('Vpc',
            CidrBlock=FindInMap("VpcCidrs", "vpc", "cidr"),
            EnableDnsSupport="true",
            EnableDnsHostnames="true",
            Tags=[
                {
                    "Key": "Name",
                    "Value": Join("-", [Ref("AWS::StackName"), "VPC"])
                }
            ]
            )
    )


def add_resource_subnet_az_1(template):
    template.add_resource(
        Subnet('PubSubnetAz1',
               DependsOn="Vpc",
               CidrBlock=FindInMap("VpcCidrs", "pubsubnet1", "cidr"),
               VpcId=Ref('Vpc'),
               AvailabilityZone=(
                   If("LambdaSupported",
                      GetAtt("AZInfo", "AZ0"),
                      Select(
                          FindInMap("AWSRegion2AZ", Ref("AWS::Region"), "AZ0"),
                          GetAZs(Ref("AWS::Region"))
                      )
                      )
               ),
               Tags=[
                   {'Key': "Name", 'Value': Join(
                    "-", [Ref("AWS::StackName"), "Subnet1"])}
               ]))


def add_resource_subnet_az_2(template):
    template.add_resource(
        Subnet('PubSubnetAz2',
               DependsOn="Vpc",
               CidrBlock=FindInMap("VpcCidrs", "pubsubnet2", "cidr"),
               VpcId=Ref('Vpc'),
               AvailabilityZone=(
                   If("LambdaSupported",
                      GetAtt("AZInfo", "AZ1"),
                      Select(
                          FindInMap("AWSRegion2AZ", Ref("AWS::Region"), "AZ1"),
                          GetAZs(Ref("AWS::Region"))
                      )
                      )
               ),
               Tags=[
                   {'Key': "Name", 'Value': Join(
                    "-", [Ref("AWS::StackName"), "Subnet2"])}
               ]))


def add_resource_subnet_az_3(template):
    template.add_resource(
        Subnet('PubSubnetAz3',
               DependsOn="Vpc",
               CidrBlock=FindInMap("VpcCidrs", "pubsubnet3", "cidr"),
               VpcId=Ref('Vpc'),
               AvailabilityZone=(
                   If("LambdaSupported",
                      GetAtt("AZInfo", "AZ2"),
                      Select(
                          FindInMap("AWSRegion2AZ", Ref("AWS::Region"), "AZ2"),
                          GetAZs(Ref("AWS::Region"))
                      )
                      )
               ),
               Tags=[
                   {'Key': "Name", 'Value': Join(
                    "-", [Ref("AWS::StackName"), "Subnet3"])}
               ]))


def add_resource_internet_gateway(template):
    template.add_resource(InternetGateway(
        "InternetGateway",
        DependsOn="Vpc",
        Tags=[
            {'Key': "Name", 'Value': Join("-", [Ref("AWS::StackName"), "IGW"])}
        ]))


def add_resource_attach_gateway(template):
    template.add_resource(VPCGatewayAttachment(
        "AttachGateway",
        DependsOn=["Vpc", "InternetGateway"],
        VpcId=Ref("Vpc"),
        InternetGatewayId=Ref("InternetGateway")))


def add_resource_route_table(template):
    template.add_resource(RouteTable(
        'RouteViaIgw',
        DependsOn="Vpc",
        VpcId=Ref('Vpc'),
        Tags=[
            {'Key': "Name", 'Value': Join("-", [Ref("AWS::StackName"), "RT"])}
        ]))


def add_resource_public_route(template):
    """
    "PublicRouteViaIgw" : {
        "DependsOn": ["AttachGateway", "RouteViaIgw"],
        "Type" : "AWS::EC2::Route",
        "Properties" : {
            "RouteTableId" : { "Ref" : "RouteViaIgw" },
            "DestinationCidrBlock" : "0.0.0.0/0",
            "GatewayId" : { "Ref" : "InternetGateway" }
        }
    },
    """
    template.add_resource(Route(
        'PublicRouteViaIgw',
        DependsOn=["AttachGateway", "RouteViaIgw"],
        RouteTableId=Ref('RouteViaIgw'),
        DestinationCidrBlock='0.0.0.0/0',
        GatewayId=Ref("InternetGateway"),
    ))


def add_resource_subnet1_route_table(template):
    """
    "PubSubnet1RouteTableAssociation" : {
        "DependsOn": ["PubSubnetAz1", "RouteViaIgw"],
        "Type" : "AWS::EC2::SubnetRouteTableAssociation",
        "Properties" : {
            "SubnetId" : { "Ref" : "PubSubnetAz1" },
            "RouteTableId" : { "Ref" : "RouteViaIgw" }
        }
    },
    """
    template.add_resource(SubnetRouteTableAssociation(
        'PubSubnet1RouteTableAssociation',
        DependsOn=["PubSubnetAz1", "RouteViaIgw"],
        SubnetId=Ref("PubSubnetAz1"),
        RouteTableId=Ref("RouteViaIgw"),
    ))


def add_resource_subnet2_route_table(template):
    """
    "PubSubnet2RouteTableAssociation" : {
        "DependsOn": ["PubSubnetAz2", "RouteViaIgw"],
        "Type" : "AWS::EC2::SubnetRouteTableAssociation",
        "Properties" : {
            "SubnetId" : { "Ref" : "PubSubnetAz2" },
            "RouteTableId" : { "Ref" : "RouteViaIgw" }
        }
    },
    """
    template.add_resource(SubnetRouteTableAssociation(
        'PubSubnet2RouteTableAssociation',
        DependsOn=["PubSubnetAz2", "RouteViaIgw"],
        SubnetId=Ref("PubSubnetAz2"),
        RouteTableId=Ref("RouteViaIgw"),
    ))


def add_resource_subnet3_route_table(template):
    """
    "PubSubnet3RouteTableAssociation" : {
        "DependsOn": ["PubSubnetAz3", "RouteViaIgw"],
        "Type" : "AWS::EC2::SubnetRouteTableAssociation",
        "Properties" : {
            "SubnetId" : { "Ref" : "PubSubnetAz3" },
            "RouteTableId" : { "Ref" : "RouteViaIgw" }
        }
    },
    """
    template.add_resource(SubnetRouteTableAssociation(
        'PubSubnet3RouteTableAssociation',
        DependsOn=["PubSubnetAz3", "RouteViaIgw"],
        SubnetId=Ref("PubSubnetAz3"),
        RouteTableId=Ref("RouteViaIgw"),
    ))

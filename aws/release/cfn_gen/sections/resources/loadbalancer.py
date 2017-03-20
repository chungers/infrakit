from troposphere import Ref, If, Join
from troposphere.elasticloadbalancing import (
    LoadBalancer, HealthCheck, ConnectionSettings, Listener)


def add_resource_external_lb(template, create_vpc, extra_listeners=None):
    """
    "ExternalLoadBalancer" : {
        "DependsOn" : ["AttachGateway", "ExternalLoadBalancerSG",
            "PubSubnetAz1", "PubSubnetAz2", "PubSubnetAz3"],
        "Type" : "AWS::ElasticLoadBalancing::LoadBalancer",
        "Properties" : {
            "ConnectionSettings" : {
                "IdleTimeout" : "600"
            },
            "Subnets": {
                "Fn::If": [
                  "HasOnly2AZs",
                  [
                          { "Ref" : "PubSubnetAz1" },
                          { "Ref" : "PubSubnetAz2" }
                  ],
                  [
                          { "Ref" : "PubSubnetAz1" },
                          { "Ref" : "PubSubnetAz2" },
                          { "Ref" : "PubSubnetAz3" }
                  ]
              ]
            },
            "Listeners" : [
                {
                    "LoadBalancerPort" : "7",
                    "InstancePort" : "7",
                    "Protocol" : "TCP"
                }
            ],
            "LoadBalancerName" : { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName"}, "ELB" ] ] },
            "CrossZone" : "true",
            "HealthCheck" : {
                "HealthyThreshold" : "2",
                "Interval" : "10",
                "Target" : "HTTP:44554/",
                "Timeout" : "2",
                "UnhealthyThreshold" : "4"
            },
            "SecurityGroups" : [ { "Ref" : "ExternalLoadBalancerSG" } ],
            "Tags": [
                {
                    "Key" : "Name",
                    "Value" : { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName"}, "ELB" ] ] }
                }
            ]
        }
    },
    """
    if create_vpc:
        depends = ["AttachGateway", "ExternalLoadBalancerSG",
                   "PubSubnetAz1", "PubSubnetAz2", "PubSubnetAz3"]
    else:
        depends = ["ExternalLoadBalancerSG"]

    listener_list = []
    listener_list.append(Listener(
        LoadBalancerPort="7",
        InstancePort="7",
        Protocol="TCP"
    ),)
    if extra_listeners:
        listener_list.extend(extra_listeners)

    template.add_resource(LoadBalancer(
        "ExternalLoadBalancer",
        DependsOn=depends,
        ConnectionSettings=ConnectionSettings(IdleTimeout=600),
        Subnets=If("HasOnly2AZs",
                   [Ref("PubSubnetAz1"), Ref("PubSubnetAz2")],
                   [Ref("PubSubnetAz1"), Ref("PubSubnetAz2"),
                    Ref("PubSubnetAz3")]),
        HealthCheck=HealthCheck(
            Target="HTTP:44554/",
            HealthyThreshold="2",
            UnhealthyThreshold="4",
            Interval="10",
            Timeout="2",
        ),
        Listeners=listener_list,
        CrossZone=True,
        SecurityGroups=[Ref("ExternalLoadBalancerSG")],
        Tags=[
            {'Key': "Name", 'Value': Join("-", [Ref("AWS::StackName"), "ELB"])}
        ]
    ))


def add_resource_ddc_ucp_lb(template, create_vpc, extra_listeners=None):
    """
        "UCPLoadBalancer": {
            "DependsOn": [
                "AttachGateway",
                "PubSubnetAz1",
                "PubSubnetAz2"
            ],
            "Properties": {
                "ConnectionSettings": {
                    "IdleTimeout": "1800"
                },
                "CrossZone": "true",
                "HealthCheck": {
                    "HealthyThreshold": "2",
                    "Interval": "10",
                    "Target": "TCP:443",
                    "Timeout": "2",
                    "UnhealthyThreshold": "4"
                },
                "Listeners": [
                    {
                        "InstancePort": "443",
                        "LoadBalancerPort": "443",
                        "Protocol": "TCP"
                    }
                ],
                "SecurityGroups": [
                    {
                        "Ref": "UCPLoadBalancerSG"
                    }
                ],
                "Subnets": [
                    {
                        "Ref": "PubSubnetAz1"
                    },
                    {
                        "Ref": "PubSubnetAz2"
                    }
                ],
                "Tags": [
                    {
                        "Key": "Name",
                        "Value": {
                            "Fn::Join": [
                                "-",
                                [
                                    {
                                        "Ref": "AWS::StackName"
                                    },
                                    "ELB-UCP"
                                ]
                            ]
                        }
                    }
                ]
            },
            "Type": "AWS::ElasticLoadBalancing::LoadBalancer"
        },
    """
    if create_vpc:
        depends = ["AttachGateway", "UCPLoadBalancerSG",
                   "PubSubnetAz1", "PubSubnetAz2", "PubSubnetAz3"]
    else:
        depends = ["UCPLoadBalancerSG"]

    listener_list = []
    listener_list.append(Listener(
        LoadBalancerPort="443",
        InstancePort="12390",
        Protocol="TCP"
    ),)
    if extra_listeners:
        listener_list.extend(extra_listeners)

    template.add_resource(LoadBalancer(
        "UCPLoadBalancer",
        DependsOn=depends,
        ConnectionSettings=ConnectionSettings(IdleTimeout=1800),
        Subnets=If("HasOnly2AZs",
                   [Ref("PubSubnetAz1"), Ref("PubSubnetAz2")],
                   [Ref("PubSubnetAz1"), Ref("PubSubnetAz2"),
                    Ref("PubSubnetAz3")]),
        HealthCheck=HealthCheck(
            Target="HTTPS:12390/_ping",
            HealthyThreshold="2",
            UnhealthyThreshold="4",
            Interval="10",
            Timeout="2",
        ),
        Listeners=listener_list,
        CrossZone=True,
        SecurityGroups=[Ref("UCPLoadBalancerSG")],
        Tags=[
            {'Key': "Name", 'Value': Join("-", [Ref("AWS::StackName"), "ELB-UCP"])}
        ]
    ))


def add_resource_ddc_dtr_lb(template, create_vpc, extra_listeners=None):
    """
        "DTRLoadBalancer": {
            "DependsOn": [
                "AttachGateway",
                "PubSubnetAz1",
                "PubSubnetAz2"
            ],
            "Properties": {
                "ConnectionSettings": {
                    "IdleTimeout": "1800"
                },
                "CrossZone": "true",
                "HealthCheck": {
                    "HealthyThreshold": "2",
                    "Interval": "10",
                    "Target": "HTTPS:8443/health",
                    "Timeout": "2",
                    "UnhealthyThreshold": "4"
                },
                "Listeners": [
                    {
                        "InstancePort": "8443",
                        "LoadBalancerPort": "443",
                        "Protocol": "TCP"
                    }
                ],
                "SecurityGroups": [
                    {
                        "Ref": "DTRLoadBalancerSG"
                    }
                ],
                "Subnets": [
                    {
                        "Ref": "PubSubnetAz1"
                    },
                    {
                        "Ref": "PubSubnetAz2"
                    }
                ],
                "Tags": [
                    {
                        "Key": "Name",
                        "Value": {
                            "Fn::Join": [
                                "-",
                                [
                                    {
                                        "Ref": "AWS::StackName"
                                    },
                                    "ELB-DTR"
                                ]
                            ]
                        }
                    }
                ]
            },
            "Type": "AWS::ElasticLoadBalancing::LoadBalancer"
        },
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
            UnhealthyThreshold="4",
            Interval="10",
            Timeout="2",
        ),
        Listeners=listener_list,
        CrossZone=True,
        SecurityGroups=[Ref("DTRLoadBalancerSG")],
        Tags=[
            {'Key': "Name", 'Value': Join("-", [Ref("AWS::StackName"), "ELB-DTR"])}
        ]
    ))

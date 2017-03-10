from troposphere import Ref, If, Join
from troposphere.elasticloadbalancing import (
    LoadBalancer, HealthCheck, ConnectionSettings, Listener)


def add_resource_external_lb(template, create_vpc):
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
        Listeners=[
            Listener(
                LoadBalancerPort="7",
                InstancePort="7",
                Protocol="TCP"
            ),
        ],
        CrossZone=True,
        SecurityGroups=[Ref("ExternalLoadBalancerSG")],
        Tags=[
            {'Key': "Name", 'Value': Join("-", [Ref("AWS::StackName"), "ELB"])}
        ]
    ))

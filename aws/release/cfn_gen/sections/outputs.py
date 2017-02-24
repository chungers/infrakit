from troposphere import Output, Ref, If, Join, GetAtt


def add_output_managers(template):
    """
    "Managers": {
        "Description": "You can see the manager nodes associated with this cluster here.
            Follow the instructions here: https://docs.docker.com/docker-for-aws/deploy/",
        "Value": {
        "Fn::Join": [ "", [
            "https://",
            { "Ref": "AWS::Region" },
            ".console.aws.amazon.com/ec2/v2/home?region=",
            { "Ref": "AWS::Region" },
            "#Instances:tag:aws:autoscaling:groupName=",
            { "Ref": "ManagerAsg" },
            ";sort=desc:dnsName"
        ]]
        }
    },
    """
    template.add_output(Output(
        "Managers",
        Description="You can see the manager nodes associated with this cluster here."
                    " Follow the instructions here: https://docs.docker.com/docker-for-aws/deploy/",
        Value=Join("", [
            "https://",
            Ref("AWS::Region"),
            ".console.aws.amazon.com/ec2/v2/home?region=",
            Ref("AWS::Region"),
            "#Instances:tag:aws:autoscaling:groupName=",
            Ref("ManagerAsg"),
            ";sort=desc:dnsName"])
    ))


def add_output_dns_target(template):
    """
    "DefaultDNSTarget" : {
        "Description" : "Use this name to update your DNS records",
        "Value" : {
            "Fn::GetAtt" : [ "ExternalLoadBalancer", "DNSName" ]
        }
    },
    """
    template.add_output(Output(
        "DefaultDNSTarget",
        Description="Use this name to update your DNS records",
        Value=GetAtt("ExternalLoadBalancer", "DNSName")
    ))


def add_output_elb_zone_id(template):
    """
    "ELBDNSZoneID": {
        "Description": "Use this zone ID to update your DNS records",
        "Value": { "Fn::GetAtt": [ "ExternalLoadBalancer", "CanonicalHostedZoneNameID" ] }
    }
    """
    template.add_output(Output(
        "ELBDNSZoneID",
        Description="Use this zone ID to update your DNS records",
        Value=GetAtt("ExternalLoadBalancer", "CanonicalHostedZoneNameID")
    ))


def add_output_az_warning(template):
    """
    "ZoneAvailabilityComment" : {
        "Description" : "Availabilty Zones Comment",
        "Value" : {
            "Fn::If" : [
              "HasOnly2AZs",
              "This region only has 2 Availabiliy Zones (AZ). If one of those AZs goes away,
               it will cause problems for your Swarm Managers. Please use a Region with at
               least 3 AZs.",
              "This region has at least 3 Availability Zones (AZ). This is ideal to ensure a
              fully functional Swarm in case you lose an AZ."
            ]
        }
    }
    """
    template.add_output(Output(
        "ZoneAvailabilityComment",
        Description="Availabilty Zones Comment",
        Value=If(
            "HasOnly2AZs",
            "This region only has 2 Availabiliy Zones (AZ). If one of those AZs goes away, "
            "it will cause problems for your Swarm Managers. Please use a Region with at "
            "least 3 AZs.",
            "This region has at least 3 Availability Zones (AZ). This is ideal to ensure a "
            "fully functional Swarm in case you lose an AZ."
        )
    ))

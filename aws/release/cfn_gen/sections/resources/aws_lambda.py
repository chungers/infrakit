# FYI lambda is a reserved word in python
# so we refer to it as aws_lambda when needed.

from troposphere import Ref, GetAtt, Join
from troposphere.logs import LogGroup
from troposphere.awslambda import Function, Code, MEMORY_VALUES

code = [
    "import cfnresponse",
    "import boto3",
    "def handler(event, context):",
    "    ec2c = boto3.client('ec2')",
    "    r = ec2c.describe_availability_zones()",
    "    azs = r.get('AvailabilityZones')",
    "    az_list = [az.get('ZoneName') for az in azs if az.get('State') == 'available']",
    "    az0 = az_list[0]",
    "    az1 = az_list[1]",
    "    if len(az_list) > 2:",
    "        az2 = az_list[2]",
    "    else:",
    "        az2 = az0",
    "    resp = {'AZ0': az0, 'AZ1': az1, 'AZ2': az2}",
    "    cfnresponse.send(event, context, cfnresponse.SUCCESS, resp)",
    "    return resp"
]


def add_resource_az_info_function(template):
    """
    "AZInfoFunction": {
      "Type": "AWS::Lambda::Function",
      "Properties": {
        "Code": {
            "ZipFile" : { "Fn::Join" : ["\n", [
                "import cfnresponse",
                "import boto3",
                "def handler(event, context):",
                "    ec2c = boto3.client('ec2')",
                "    r = ec2c.describe_availability_zones()",
                "    azs = r.get('AvailabilityZones')",
                "    az_list = [az.get('ZoneName') for az in azs if az.get('State') == 'available']",
                "    az0 = az_list[0]",
                "    az1 = az_list[1]",
                "    if len(az_list) > 2:",
                "        az2 = az_list[2]",
                "    else:",
                "        az2 = az0",
                "    resp = {'AZ0': az0, 'AZ1': az1, 'AZ2': az2}",
                "    cfnresponse.send(event, context, cfnresponse.SUCCESS, resp)",
                "    return resp"
            ]]}
        },
        "Handler": "index.handler",
        "Role": { "Fn::GetAtt" : ["LambdaExecutionRole", "Arn"] },
        "Runtime": "python2.7",
        "Timeout": "10"
      }
    },
    """
    template.add_resource(Function(
        "AZInfoFunction",
        Condition="LambdaSupported",
        Code=Code(
            ZipFile=Join("\n", code)
        ),
        Handler="index.handler",
        Role=GetAtt("LambdaExecutionRole", "Arn"),
        Runtime="python2.7",
        MemorySize="128",
        Timeout="10"
    ))

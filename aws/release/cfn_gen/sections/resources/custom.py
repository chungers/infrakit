from troposphere import Ref, GetAtt
from troposphere.cloudformation import AWSCustomObject


class CustomAZInfo(AWSCustomObject):
    resource_type = "Custom::AZInfo"

    props = {
        'ServiceToken': (basestring, True),
        'Region': (basestring, True)
    }


def add_resource_custom_az_info(template):
    """
    "AZInfo": {
      "Type": "Custom::AZInfo",
      "Properties": {
        "ServiceToken": { "Fn::GetAtt" : ["AZInfoFunction", "Arn"] },
        "Region": { "Ref": "AWS::Region" }
      }
    """
    template.add_resource(CustomAZInfo(
        "AZInfo",
        Condition="LambdaSupported",
        ServiceToken=GetAtt("AZInfoFunction", "Arn"),
        Region=Ref("AWS::Region"))
    )

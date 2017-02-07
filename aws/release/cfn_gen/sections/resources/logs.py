from troposphere import Ref, Join
from troposphere.logs import LogGroup


def add_resource_log_group(template):
    """
    "DockerLogGroup": {
      "Type" : "AWS::Logs::LogGroup",
      "Condition" : "CreateLogResources",
      "Properties" : {
        "LogGroupName" : { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName"}, "lg" ] ] },
        "RetentionInDays" : 7
      }
    },
    """
    template.add_resource(LogGroup(
        "DockerLogGroup",
        Condition="CreateLogResources",
        LogGroupName=Join("-", [Ref("AWS::StackName"), "lg"]),
        RetentionInDays=7)
    )

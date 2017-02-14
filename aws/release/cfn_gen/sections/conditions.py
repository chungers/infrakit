from troposphere import Equals, Or, FindInMap, Ref


def add_condition_create_log_resources(template):
    template.add_condition(
        "CreateLogResources",
        Equals(Ref("EnableCloudWatchLogs"), "yes")
        )


def add_condition_hasonly2AZs(template):
    template.add_condition(
        "HasOnly2AZs",
        Equals(
            FindInMap("AWSRegion2AZ", Ref("AWS::Region"), "NumAZs"),
            "2")
        )


def add_condition_EFSSupported(template):
    template.add_condition(
        "EFSSupported",
        Or(
            Equals("us-west-2", Ref("AWS::Region")),
            Equals("us-east-1", Ref("AWS::Region")),
            Equals("eu-west-1", Ref("AWS::Region"))
        )
    )

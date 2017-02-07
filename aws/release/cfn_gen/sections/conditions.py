from troposphere import Equals, FindInMap, Ref


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

from troposphere import Equals, FindInMap, Ref, And, Condition


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


def add_condition_CloudStorEFS_selected(template):
    template.add_condition(
        "CloudStorEfsSelected",
        Equals(Ref("EnableCloudStorEfs"), "yes")
        )


def add_condition_EFSSupported(template):
    template.add_condition(
        "EFSSupported",
        Equals(
            FindInMap("AWSRegion2AZ", Ref("AWS::Region"), "EFSSupport"),
            "yes")
    )


def add_condition_InstallCloudStorEFSPreReqs(template):
    template.add_condition(
        "InstallCloudStorEFSPreReqs",
        And(
            Condition("EFSSupported"),
            Condition("CloudStorEfsSelected"),
        )
    )


def add_condition_LambdaSupported(template):
    template.add_condition(
        "LambdaSupported",
        Equals(
            FindInMap("AWSRegion2AZ", Ref("AWS::Region"), "LambdaSupport"),
            "yes")
    )


def add_condition_EBSOptimized(template):
    template.add_condition(
        "EBSOptimized",
        Equals(Ref("EnableEbsOptimized"), "yes")
    )

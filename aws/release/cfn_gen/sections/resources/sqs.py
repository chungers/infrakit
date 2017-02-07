from troposphere.sqs import Queue


def add_resources_sqs_cleanup(template):
    """
    "SwarmSQSCleanup" : {
        "Type": "AWS::SQS::Queue",
        "Properties": {
            "ReceiveMessageWaitTimeSeconds": 10,
            "MessageRetentionPeriod": 43200
        }
    },
    """
    template.add_resource(Queue(
        "SwarmSQSCleanup",
        ReceiveMessageWaitTimeSeconds=10,
        MessageRetentionPeriod=43200
    ))


def add_resources_sqs_swarm(template):
    """
    "SwarmSQS" : {
        "Type": "AWS::SQS::Queue",
        "Properties": {
            "ReceiveMessageWaitTimeSeconds": 10,
            "MessageRetentionPeriod": 43200
        }
    },
    """
    template.add_resource(Queue(
        "SwarmSQS",
        ReceiveMessageWaitTimeSeconds=10,
        MessageRetentionPeriod=43200
    ))

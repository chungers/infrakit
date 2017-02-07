from troposphere import Ref, Join
from troposphere.dynamodb import (KeySchema, AttributeDefinition,
                                  ProvisionedThroughput, Table)


def add_resource_dyn_table(template):
    """
    "SwarmDynDBTable" : {
        "DependsOn" : "ExternalLoadBalancer",
          "Type" : "AWS::DynamoDB::Table",
          "Properties" : {
            "TableName": { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName"}, "dyndbtable" ] ] },
            "AttributeDefinitions": [ {
              "AttributeName" : "node_type",
              "AttributeType" : "S"
            } ],
            "KeySchema": [
              { "AttributeName": "node_type" , "KeyType": "HASH" }
            ],
            "ProvisionedThroughput" : {
              "ReadCapacityUnits" : 1,
              "WriteCapacityUnits" : 1
            }
          }
      },
    """
    template.add_resource(Table(
        "SwarmDynDBTable",
        DependsOn="ExternalLoadBalancer",
        TableName=Join("-", [Ref("AWS::StackName"), "dyndbtable"]),
        AttributeDefinitions=[
            AttributeDefinition(
                AttributeName="node_type",
                AttributeType="S"
            ),
        ],
        KeySchema=[
            KeySchema(
                AttributeName="node_type",
                KeyType="HASH"
            )
        ],
        ProvisionedThroughput=ProvisionedThroughput(
            ReadCapacityUnits=1,
            WriteCapacityUnits=1
        )
    ))

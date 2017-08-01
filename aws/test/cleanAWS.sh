#!/bin/bash

# Script to delete any EC2 Keys that might not have been deleted
# Stack is deleted directly from Makefile, because name is assigned
# in Makefile

#Get all the key names that have QA_TEST in their name
KEYNAMES=$(aws ec2 describe-key-pairs | jq '.KeyPairs[] | select(.KeyName | contains("QA_TEST"))' | jq .KeyName | tr ' ' '\n' )


# Delete all the key pairs which might be lingering
if [ -z "$KEYNAMES" ];
      then
      echo "No Keys to clean up"
else
        while read -r KEYNAME; do
                KEYNAME=$(echo $KEYNAME | sed -e 's/\"//' | sed -e 's/\"//')
                echo "aws ec2 delete-key-pair --key-name  $KEYNAME"
                aws ec2 delete-key-pair --key-name  "$KEYNAME"
        done <<< "$KEYNAMES"
fi

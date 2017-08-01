#!/bin/bash

AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_DEFAULT_REGION=
CLEAN="false"
NAME=
MANCOUT=
WORKERCOUNT=
TEMPLATE_URL=
KEYNAME="QA_TEST_""$(( 1 + $RANDOM % 200 ))"

usage() {
  echo
  echo "Usage createStack.sh [OPTIONS]"
  echo
  echo
  echo "Options: "
  echo
  echo "Flag    Info            Description"
  echo "-c      Clean           Specify if you want the stack deleted"
  echo "-h      Help            Prints a usage statement with additional details"
  echo "-i      Access Key      Specify the AWS Access Key ID"
  echo "-m      Manager count   Specify number of managers"
  echo "-n      Stack name      Specify the name for the stack"
  echo "-p      Port            Specify the port you want to SSH into"
  echo "-r      Region          Specify the region of your stack"
  echo "-s      Secret Key      Specify the AWS Secret Access Key"
  echo "-t      Template URL    Specify the AWS template URL"
  echo "-w      Worker count    Specify the worker count"
}


while getopts ":i:n:m:p:r:s:t:w:ch" opt; do
  case $opt in
    c)
      CLEAN="true"
      ;;
    h)
      usage
      ;;
    i)
      AWS_ACCESS_KEY_ID=$OPTARG
      ;;
    m)
      MANCOUNT=$OPTARG
      ;;
    n)
      NAME=$OPTARG
      ;;
    p)
      PORT=$OPTARG
      ;;
    r)
      AWS_DEFAULT_REGION=$OPTARG
      ;;
    s)
      AWS_SECRET_ACCESS_KEY=$OPTARG
      ;;
    t)
      TEMPLATE_URL=$OPTARG
      ;;
    w)
      WORKERCOUNT=$OPTARG
      ;;
    :)
      "Option -$OPTARG requires an argument"
      exit
      ;;
    ?)
      echo "Invalid option: -$OPTARG"
      usage
      exit
      ;;
  esac

  # Catches the error where an argument is required following
  # the flag but it is ommited hence the next flag is used as the
  # argument.
  #  ex. createStack.sh -t -r    (the parameter for -t would be interpreted as
  #  -r but should be a template)
  if [ "$(echo "$OPTARG" | grep "^-")" != "" ]; then
    echo "no argument specified or hyphen placed at beginning of arg: $OPTARG"
    exit
  fi

done



# Sub routine used to retry and wait on different functions until they succeeded
# or there no more attempts stack_creation: Checks if the stack was created with 
# status: CREATE_COMPLETE 
# scp: Copies a helper script to stack (used to check how many workers
# and managers there are)
#
# ssh: Accesses stack and checks that all managers and 
# workers are up
#
#RETRIES WAIT CMD
wait() {
  for (( i = 1; i <= $1; i++ ))
  do
    if [ "$3" == "stack_creation" ];
        then
        FORMATIONSTATUS=$(aws cloudformation describe-stacks --stack-name $STACK_ID | jq .Stacks[0].StackStatus)
         if [ "$FORMATIONSTATUS"  == "\"CREATE_COMPLETE\"" ]; 
               then 
               echo "STACK CREATION COMPLETE"  
               break
           elif [[ "$FORMATIONSTATUS" == "ROLLBACK_FAILED" || "$FORMATIONSTATUS" == "ROLLBACK_COMPLETE" ||  "$FORMATIONSTATUS" == "ROLLBACK_IN_PROGRESS" || "$FORMATIONSTATUS" == "CREATE_FAILED" ]]
               then
               echo "$FORMATIONSTATUS"
               exit
            fi

    elif  [ $3 == "scp" ];
        then scp -o StrictHostKeyChecking=no /usr/bin/swarmCountCheck.sh docker@$ip:/home/docker
        r=$?
    elif  [ $3 == "ssh" ];
        then ssh -o StrictHostKeyChecking=no  docker@$ip "bash swarmCountCheck.sh $MANCOUNT $WORKERCOUNT"
        r=$?
    fi

    #see if check failed, and sleep otherwise
    if [ $i -eq $1 ];
        then echo "$3 failed"
    elif [[ ($3 == "ssh" || $3 == "scp") && $r -eq 0 ]];
        then break
    else
        sleep $2
    fi

  done
}



#Place the id and secret in a config file
mkdir -p ~/.aws
echo "[default]" > ~/.aws/credentials
echo "aws_secret_access_key=$AWS_SECRET_ACCESS_KEY" >> ~/.aws/credentials
echo "aws_access_key_id=$AWS_ACCESS_KEY_ID" >> ~/.aws/credentials


#Exporting environment variables for AWS region  
export AWS_DEFAULT_REGION=$AWS_DEFAULT_REGION

#Generate SSH Key
ssh-keygen -N '' -f /root/.ssh/id_rsa  -t rsa -b 2048
SSHPATH="file://~/.ssh/id_rsa.pub"

#copy ssh key into AWS  Make sure to delete this key after the test finishes
aws ec2 --region=$AWS_DEFAULT_REGION import-key-pair --key-name $KEYNAME --public-key-material $SSHPATH 


#Create AWS Stack from template, and save stack id
STACK_ID=$(aws cloudformation create-stack --template-url $TEMPLATE_URL --parameters ParameterKey=ClusterSize,ParameterValue=$WORKERCOUNT ParameterKey=EnableCloudWatchLogs,ParameterValue=no ParameterKey=EnableCloudWatchLogs,ParameterValue=no ParameterKey=InstanceType,ParameterValue=t2.micro ParameterKey=KeyName,ParameterValue=$KEYNAME ParameterKey=ManagerInstanceType,ParameterValue=t2.micro ParameterKey=ManagerSize,ParameterValue=$MANCOUNT --capabilities CAPABILITY_IAM --stack-name $NAME | jq .StackId  | sed -e 's/\"//' | sed -e 's/\"//')

echo $STACK_ID

#Check that the stack creation is successful / complete 
wait 40  30s stack_creation

# Get any of the manager ip addresses for  sshing
# Worse case scenario all ip addresses come up as None
# Might be worth adding a check or looping
ip=$(aws ec2 describe-instances --filters "Name=tag:Name,Values=$NAME-Manager"  --query "Reservations[*].Instances[*].[PublicIpAddress]" --output=text | grep [0-9] | head -1)

echo IP address: $ip

#SCP swarmCountCheck.sh script (helper script that ensures manager and workers are all up)
wait 10 2s scp

#SSH and run swarmCountCheck.sh script
wait 10 60s ssh

#Run set of e2e tests
ssh -o StrictHostKeyChecking=no  docker@$ip "docker run -v /var/run/docker.sock:/var/run/docker.sock dockere2e/tests"

#################### 
##   Clean up     ##
#####################
#Delete ssh public key that was uploaded to AWS
aws ec2 delete-key-pair --key-name $KEYNAME

#Delete AWS Stack
if [ "$CLEAN" == "true" ]; then
  echo "Deleting the stack"
  aws cloudformation delete-stack --stack-name "$STACK_ID"
fi

#Delete aws config file
rm -f ~/.aws/credentials

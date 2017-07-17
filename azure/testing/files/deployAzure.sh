#!/bin/bash
#Guideline:

#1. Generate SSH key pair
#2. Create resource group in azure and deploy swarm
#3. Wait for resource group to succeed
#4  Parse out IP address of manager from deployment output of resource group
#5  SSH into manager (may not be ready right away - continually check)
#6  Run a bunch of tests from a container (e.g. create swarm services, create storage, etc)

RESOURCEGROUP=
WORKERCOUNT=
LOC=
TENANT=
PORT=
MANCOUNT=
SPSECRET=
SPID=
TEMPLATEURI=
CLEAN="false"


#1 Generate SSH key
ssh-keygen -N '' -f ~/.ssh/id_rsa  -t rsa -b 2048
SSH=$(cat ~/.ssh/id_rsa.pub) 

#Timeout set for checking if deployment succeeded
TIMEOUT=3600

usage() {
    echo  
    echo "Usage: deployAzure.sh  [OPTIONS]"
    echo
    echo
    echo
    echo "Options: "
    echo "Flag    Info               Description"
    echo
    echo "-c      clean              Deletes the resource group"
    echo "-h      help               Print Usage"
    echo "-i      SPID               Specify the Service Principal App ID"
    echo "-l      location           Specify the location"
    echo "-m      managers           Specify the number of managers"
    echo "-n      tenant             Specify the tenant id"
    echo "-p      Port               The port you want to connect on"
    echo "-r      resourcegroup      Specify the resourcegroup"
    echo "-s      SPSECRET           Specify the Service Principal App Secret"
    echo "-t      templateURI        Specify the template"
    echo "-w      workers            Specify the number of workers"   
}

#Sub routine used to SSH or SCP takes in RETRIES, WAIT, and the CMD (SCP or SSH)
#RETRIES WAIT CMD
wait() {

for (( i = 1; i <= $1; i++ ))
do
        if [ "$3" == "ssh" ];
                then
                ssh -o StrictHostKeyChecking=no -p $PORT docker@$ip "bash /home/docker/swarmCountCheck.sh  $MANCOUNT $WORKERCOUNT"
                r=$?
        elif  [ $3 == "scp" ];
                then
                scp -P $PORT  -o StrictHostKeyChecking=no /usr/bin/swarmCountCheck.sh docker@$ip:/home/docker
                r=$?
        fi
  

        if [ $r -eq 0 ];
                then break
        elif [ $i -eq $1 ];
                then echo "$3 failed"
                exit
        else
                sleep $2 
        fi

done
}


#Parameter Parsing
while getopts ":i:l:n:m:p:r:s:t:w:hc" opt; do
	case $opt in
    	r)
		RESOURCEGROUP=$OPTARG
	    	;;
    	w)
	    	WORKERCOUNT=$OPTARG
	    	;;
    	l)
	    	LOC=$OPTARG
	    	;;
    	n)
      	   	TENANT=$OPTARG
           	;;
   	p)
           	PORT=$OPTARG
           	;;
    	m)
	   	MANCOUNT=$OPTARG
	   	;;
    	h)
           	usage
           	exit
           	;;
    	c)
           	CLEAN="true"
           	;;
    	s)
           	SPSECRET=$OPTARG
           	;;
    	i)
           	SPID=$OPTARG
           	;;
    	t)
	   	TEMPLATEURI=$OPTARG
	   	;;
    	:)
	   	echo "Option -$OPTARG requires an argument." 
	   	exit
	   	;;
    	\?)
	   	echo "Invalid option: -$OPTARG"
           	usage
	   	exit
	   	;;
  	esac
     
        # Catches the error where an argument is required following
        # the flag but it is omitted hence the next flag is used as the
        # argument.  
        #  ex. deployAzure.sh -l -r    (the parameter for -l would be interpreted as
        #  -r but should be a location) 
        if [ "$(echo "$OPTARG" | grep "^-")" != "" ]; then
                echo "no argument specified or hyphen placed at beginning of arg: $OPTARG"
                exit
        fi

done


#Login in to azure with service principal and secret
az login --service-principal -u "https://edition-test-sp" -p $SPSECRET --tenant $TENANT 


#Create resource group and store the id (used later to obtain ip  address)
id=$(az group create --name $RESOURCEGROUP  --location $LOC | jq .id  | sed -e 's/\"//' | sed -e 's/\"//')

#Deploy swarm in resource group
#Use template URI or file depending on what user specified
if [[ $(echo "$TEMPLATEURI" | grep http) != "" ]]; then
        az group deployment create --name docker.template --resource-group $RESOURCEGROUP --template-uri $TEMPLATEURI  --parameters "{\"adServicePrincipalAppID\":{\"value\": \"$SPID\" }, \"adServicePrincipalAppSecret\":{\"value\": \"$SPSECRET\"}, \"sshPublicKey\": {\"value\": \"$SSH\"},  \"managerCount\": {\"value\": $MANCOUNT}, \"workerCount\": {\"value\": $WORKERCOUNT}}" 
else
        az group deployment create --name docker.template --resource-group $RESOURCEGROUP --template-file $TEMPLATEURI  --parameters "{\"adServicePrincipalAppID\":{\"value\": \"$SPID\" }, \"adServicePrincipalAppSecret\":{\"value\": \"$SPSECRET\"}, \"sshPublicKey\": {\"value\": \"$SSH\"},  \"managerCount\": {\"value\": $MANCOUNT}, \"workerCount\": {\"value\": $WORKERCOUNT}}" 
fi

#Wait until created with provisioningState at Succeeded 
az group deployment wait --name docker.template --resource-group $RESOURCEGROUP --created --timeout $TIMEOUT 

#Get the IP to ssh into the deployment
#Use the id to get the ip address
ip=$(az resource show --id $id/providers/Microsoft.Network/publicIPAddresses/dockerswarm-externalSSHLoadBalancer-public-ip | jq .properties.ipAddress | sed -e 's/\"//'| sed -e 's/\"//' )


#SCP  a script that verifies the manager and worker counts
# and then run the script from the container as opposed to running
#the script directly, use a loop to attempt to scp multiple times if necessary
wait 10 2s scp

#Verify that the current number of workers, and managers are up using
#swarmCountCheck.sh script looping a few times if necessary to wait for all
#resources to come up
wait 10 60s ssh


#Pull and run a docker container with a few tests
#Official version of the end to end test from docker hub
ssh -o StrictHostKeyChecking=no -p $PORT docker@$ip '/usr/local/bin/docker run -v /var/run/docker.sock:/var/run/docker.sock dockere2e/tests'

#If the clean flag was specified delete the resource group that was created
if [ "$CLEAN" == "true" ]; then	
        echo "Deleting Resource Group..."
        az group delete --name $RESOURCEGROUP -y
fi

# Azure End to End Testing

### Overview
The Makefile is intended to execute a full end to end test that is incorporated with a Jenkins build. 
This process involves the following steps 

1. Obtain the VHD url, and a current template (from the Jenkins job)
2. Update the VHD url in the current template
3. Log into azure, create a resource group, and deploy the template
4. SSH into the swarm and run a set of tests

The main driver of all this is the deployAzure.sh script. 


### Creating the Current Template
As mentioned in the [editions/azure](https://github.com/docker/editions/blob/master/azure/README.md) README the Marketplace
portal approval process takes time. Instead of waiting for a template to be published you can launch Docker from Azure
directly from a VHD in a storage account.

The dockerfile.azuretemplate contains the stg_account_arm_template.py file which takes an initial template, a VHD_URL, and output file 
as parameters, and creates a template that has the VHD url in the corresponding places, which allows Docker for Azure to be launched using a VHD.

The template, and the VHD url are obtained from the Jenkins job.

1. VHD URL: s3://docker-ci-editions/vhd/`MOBY_COMMIT`/vhd_blob_url.out
2. Template: https: //editions-stage-us-east-1-150610-005505.s3.amazonaws.com/azure/`CHANNEL`/`MOBY_COMMIT`/Docker.tmpl


### Deploying the Template
The template is deployed by using the `AZ CLI`, which is installed in a Docker container maintained by Microsoft, and
has a few extra files added used to automatically deploy the template. 

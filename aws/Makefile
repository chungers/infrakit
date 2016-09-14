USER := default
STACK := editionstest
STACK_NAME := $(USER)-$(STACK)
TEMPLATE_URL := https://docker-for-aws.s3.amazonaws.com/aws/beta/aws-v1.12.1-beta5.json
KEYNAME := defaultkey
INSTANCE_TYPE := t2.micro
MANAGER_INSTANCE_TYPE := t2.small
CLUSTER_SIZE := 1
DEV := 0

ifeq ($(DEV),0)
TEMPLATE := --template-url $(TEMPLATE_URL)
else
TEMPLATE := --template-body file:///$(shell pwd)/cloudformation/docker_for_aws.json
endif

default:
	aws cloudformation create-stack \
		--capabilities CAPABILITY_IAM \
		--stack-name "$(STACK_NAME)" \
		$(TEMPLATE) \
		--parameters \
		"ParameterKey=KeyName,ParameterValue=$(KEYNAME)" \
		"ParameterKey=InstanceType,ParameterValue=$(INSTANCE_TYPE)" \
		"ParameterKey=ManagerInstanceType,ParameterValue=$(MANAGER_INSTANCE_TYPE)" \
		"ParameterKey=ClusterSize,ParameterValue=$(CLUSTER_SIZE)"

clean:
	aws cloudformation delete-stack --stack-name "$(STACK_NAME)"

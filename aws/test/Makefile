#MAKEFILE used to deploy AWS end to end test

#Run continaer based on docker4x/awscli with some additional files that
#create an aws stack from a template and then run a set of end to end tests


E2E_CONTAINER := aws_e2e_test
TEMPLATE_PREFIX := https://editions-stage-us-east-1-150610-005505.s3.amazonaws.com/aws
STACK_FILE := stackName.tmp

ifeq ($(EDITIONS_COMMIT), )
	EDITIONS_COMMIT := e61493fecc99ebd99668517fed7294b82f6da666
endif

ifeq ($(BUILDNUMBER), )
	BUILDNUMBER := 353
endif

ifeq ($(CHANNEL), )
	CHANNEL := test
endif

#Added time stamp so you should never have duplicate stack names
ifeq ($(STACKNAME), )
	STACKNAME := testing$(shell date +%s)
endif

ifeq ($(MANCOUNT), )
	MANCOUNT := 1
endif

ifeq ($(WORKERCOUNT), )
	WORKERCOUNT := 1
endif

ifeq ($(REGION), )
	REGION := us-east-2
endif


#Suffix would be used to change what template is used
# i.e: ee, ce, cloud, no-vpc, etc.
ifeq ($(TEMPLATE_TYPE),)
	TEMPLATE_TYPE := Docker.tmpl
endif

ifeq ($(TEMPLATE_URL),  )
	TEMPLATE_URL := $(TEMPLATE_PREFIX)/$(CHANNEL)/$(EDITIONS_COMMIT)/$(TEMPLATE_TYPE)
endif

default: checks
	echo "$(STACKNAME)" >> "$(STACK_FILE)" 
	docker build . --tag $(E2E_CONTAINER)
	docker run -ti \
	-v /var/run/docker.sock:/var/run/docker.sock \
	-v /usr/bin/docker:/usr/bin/docker \
	$(E2E_CONTAINER) \
	-i  "$(AWS_ACCESS_KEY_ID)" \
	-m "$(MANCOUNT)" \
	-n "$(STACKNAME)" \
	-r "$(REGION)" \
	-s "$(AWS_SECRET_ACCESS_KEY)" \
	-t "$(TEMPLATE_URL)" \
	-w "$(WORKERCOUNT)" \ 
	-c
	$(MAKE) clean


# Check to make sure the ACCESS KEY ID and the SECRET ACCESS KEY were specified
checks:
ifeq ($(AWS_ACCESS_KEY_ID), )
	$(error "AWS_ACCCESS_KEY_ID is not set")
endif
ifeq ($(AWS_SECRET_ACCESS_KEY), )
	$(error "AWS_SECRET_ACCESS_KEY is not set")
endif

# Deletes containers that were made
# Deletes stacks or keypairs that might be lingering
# if the deployment failed midway
clean:
	docker rmi -f $(E2E_CONTAINER)
	bash cleanAWS.sh 
	aws cloudformation delete-stack --stack-name "$(shell cat "$(STACK_FILE)")"
	rm "$(STACK_FILE)"
	rm ~/.aws/credentials

USER := default
STACK := editionstest
STACK_NAME := $(USER)-$(STACK)
TEMPLATE_URL := https://docker-for-aws.s3.amazonaws.com/aws/beta/latest.json
KEYNAME := defaultkey
INSTANCE_TYPE := t2.micro
MANAGER_INSTANCE_TYPE := t2.small
CLUSTER_SIZE := 1
DEV := 0

ifeq ($(ROOT_DIR),)
	ROOT_DIR := ${CURDIR}
endif

ifeq ($(BUILD),)
	BUILD := 1
endif

AMI_S3_BUCKET := docker-ci-editions
AMI_S3_PATH := ami
AMI_SRC_REGION := us-west-2

STAGE_S3_BUCKET := editions-stage-us-east-1-150610-005505

ifneq ($(MOBY_COMMIT),)
	AMI_S3_PATH := $(AMI_S3_PATH)/$(MOBY_COMMIT)
endif

ifeq ($(RELEASE),1)
	MAKE_AMI_PUBLIC = yes
endif

ifeq ($(AMI_ID),)
	EDITIONS_META := $(shell docker run --rm docker4x/awscli:latest s3api --no-sign-request get-object --bucket $(AMI_S3_BUCKET) --key $(AMI_S3_PATH)/ami_id.out docker.out | jq -r '.Metadata' > metadata.json )
	EDITIONS_VERSION := $(shell jq -r '.editions_version' metadata.json)
	DOCKER_VERSION := $(shell jq -r '.docker_version' metadata.json)
	MOBY_COMMIT := $(shell jq -r '.moby_commit' metadata.json)
	CHANNEL := $(shell jq -r '.channel' metadata.json)
	AMI_ID := $(shell docker run --rm -v /tmp:/tmp docker4x/awscli:latest s3 --quiet --no-sign-request cp s3://${AMI_S3_BUCKET}/${AMI_S3_PATH}/ami_id.out /tmp/ami_id.out && cat /tmp/ami_id.out)
	AWS_TARGET_PATH = dist/aws/${CHANNEL}/${EDITIONS_VERSION}
endif

DOCKER_AWS_ACCOUNT_URL := https://s3.amazonaws.com/docker-for-aws/data/docker_accounts.txt

ifeq ($(DEV),0)
TEMPLATE := --template-url $(TEMPLATE_URL)
else
TEMPLATE := --template-body file:///$(shell pwd)/cloudformation/docker_for_aws.json
endif

TOOLS := docker4x/tools:latest
TOOLSRUN := docker run --rm -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY $(TOOLS) python

CFNGEN := docker4x/cfngen:latest

export

.PHONY: release

deploy:
	aws cloudformation create-stack \
		--capabilities CAPABILITY_IAM \
		--stack-name "$(STACK_NAME)" \
		$(TEMPLATE) \
		--parameters \
		"ParameterKey=KeyName,ParameterValue=$(KEYNAME)" \
		"ParameterKey=InstanceType,ParameterValue=$(INSTANCE_TYPE)" \
		"ParameterKey=ManagerInstanceType,ParameterValue=$(MANAGER_INSTANCE_TYPE)" \
		"ParameterKey=ClusterSize,ParameterValue=$(CLUSTER_SIZE)"

clean-stack:
	aws cloudformation delete-stack --stack-name "$(STACK_NAME)"

release: buildtools
	@echo "\033[0;1;34m++ Create AWS release of EDITIONS_VERSION: ${EDITIONS_VERSION} on CHANNEL: ${CHANNEL}\033[0m"
	docker run --rm \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e DOCKER_VERSION="${DOCKER_VERSION}" \
		-e EDITIONS_VERSION="${EDITIONS_VERSION}" \
		-e AMI_ID="${AMI_ID}" \
		-e AMI_SRC_REGION="${AMI_SRC_REGION}" \
		-e CHANNEL="${CHANNEL}" \
		-e CHANNEL_CLOUD="${CHANNEL}-cloud" \
		-e AWS_ACCOUNT_LIST_URL="${DOCKER_AWS_ACCOUNT_URL}" \
		-e MAKE_AMI_PUBLIC="${MAKE_AMI_PUBLIC}" \
		-e FLAGS="--upload" \
		-e UPLOAD_S3_BUCKET="${STAGE_S3_BUCKET}" \
		-e UPLOAD_S3_KEY="${AWS_STAGE_ACCESS_KEY_ID}" \
		-e UPLOAD_S3_SECRET="${AWS_STAGE_SECRET_ACCESS_KEY}" \
		-e MOBY_COMMIT=${MOBY_COMMIT}/${BUILD_NUMBER} \
		-v ${CURDIR}/cfn_gen/outputs:/home/docker/outputs \
		docker4x/cfngen

nightly:
	@echo "\033[0;1;34m++ Start nightly release\033[0m"
	$(MAKE) buildtools
	MAKE_AMI_PUBLIC=no $(MAKE) release
	@echo "\033[0;1;34m++ Run cleanup, remove things that are more than X days old.\033[0m"
	$(TOOLSRUN) cleanup.py
	@echo "\033[0;1;34m++ sleep to help with API throttle limits\033[0m"
	sleep 60
	$(MAKE) test
	@echo "\033[0;1;34m++ Rebuild the nightly index page.\033[0m"
	$(TOOLSRUN) build_index.py

template: buildtools
	@echo "\033[0;1;34m++ Create AWS template of EDITIONS_VERSION: ${EDITIONS_VERSION} on CHANNEL: ${CHANNEL}\033[0m"
	docker run \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e DOCKER_VERSION="${DOCKER_VERSION}" \
		-e EDITIONS_VERSION="${EDITIONS_VERSION}" \
		-e CHANNEL="${CHANNEL}" \
		-e CHANNEL_CLOUD="${CHANNEL}-cloud" \
		-e MOBY_COMMIT \
		-e FLAGS="--template" \
		-v ${ROOT_DIR}/dist:/home/docker/dist \
		docker4x/cfngen

buildtools: clean
	-mkdir -p tmp/ $(ROOT_DIR)/dist/
	@echo "\033[0;1;34m+++ Build cfngen image\033[0m"
	docker build -t $(CFNGEN) -f release/cfn_gen/Dockerfile release/cfn_gen
	@echo "\033[0;1;34m+++ Build common tools image\033[0m"
	docker build -t $(TOOLS) -f release/Dockerfile.tools .

test: buildtools
	@echo "\033[0;1;34m++ Run tests\033[0m"
	$(TOOLSRUN) test/cfn.py -c https://docker-for-aws.s3.amazonaws.com/aws/nightly/latest.json -f results -t oss
	@echo "\033[0;1;34m++ Sleep to help with API throttle limits\033[0m"
	sleep 60
	@echo "\033[0;1;34m++ Run Cloud tests\033[0m"
	$(TOOLSRUN) test/cfn.py -c https://docker-for-aws.s3.amazonaws.com/aws/cloud-nightly/latest.json -f cloud_results -t cloud

notify:
	@echo "\033[0;1;34m++ Notify results\033[0m"
	$(TOOLSRUN) notify.py

clean:
	@echo "\033[0;1;34m++ Clean existing artifacts\033[0m"
	-docker rmi -f $(TOOLS) $(CFNGEN)

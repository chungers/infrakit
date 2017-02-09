from troposphere import Template, Base64, Join

from sections import mappings
from sections import parameters
from sections import metadata
from sections import conditions
from sections import outputs
from sections import resources


class AWSBaseTemplate(object):

    def __init__(self, docker_version, edition_version, channel, amis,
                 create_vpc=True, template_description=None):
        self.template = Template()
        self.parameters = {}
        self.parameter_labels = {}
        self.docker_version = docker_version
        self.edition_version = edition_version
        self.channel = channel
        self.amis = amis
        self.create_vpc = create_vpc
        self.template_description = template_description

        flat_edition_version = edition_version.replace(" ", "").replace("_", "").replace("-", "")
        self.flat_edition_version = flat_edition_version
        flat_edition_version_upper = flat_edition_version.capitalize()
        self.flat_edition_version_upper = flat_edition_version_upper
        self.docker_for_aws_version = u"aws-v{}-{}".format(
            self.docker_version, self.flat_edition_version)

    def build(self):
        self.add_template_version()
        self.add_template_description()
        self.add_paramaters()
        self.add_conditions()
        self.add_metadata()
        self.add_mappings()
        self.add_resources()
        self.add_outputs()

    def add_template_version(self):
        self.template.add_version('2010-09-09')

    def add_template_description(self):
        if self.template_description:
            description = self.template_description
        else:
            description = u"Docker for AWS {} ({})".format(
                self.docker_version, self.edition_version)
        self.template.add_description(description)

    def get_parameter_groups(self):
        return metadata.parameter_groups()

    def parameter_groups(self):
        """ Override to add more parameters """
        return self.get_parameter_groups()

    def add_metadata(self):
        metadata.metadata(self.template, self.parameter_groups(), self.parameter_labels)

    def add_conditions(self):
        conditions.add_condition_create_log_resources(self.template)
        conditions.add_condition_hasonly2AZs(self.template)

    def add_mappings(self):
        mappings.add_mapping_aws2az(self.template)
        mappings.add_mapping_version(
            self.template, self.docker_version, self.docker_for_aws_version, self.channel)
        mappings.add_mapping_vpc_cidrs(self.template)
        mappings.add_mapping_instance_type_2_arch(self.template)
        mappings.add_mapping_amis(self.template, self.amis)

    def add_to_parameters(self, result):
        key, value = result
        self.parameter_labels[key] = value

    def add_paramaters(self):
        self.add_to_parameters(parameters.add_parameter_keyname(self.template))
        self.add_to_parameters(parameters.add_parameter_instancetype(self.template))
        self.add_to_parameters(parameters.add_parameter_manager_instancetype(self.template))

        self.add_to_parameters(parameters.add_parameter_cluster_size(self.template))
        self.add_to_parameters(parameters.add_parameter_worker_disk_size(self.template))
        self.add_to_parameters(parameters.add_parameter_worker_disk_type(self.template))

        self.add_to_parameters(parameters.add_parameter_manager_size(self.template))
        self.add_to_parameters(parameters.add_parameter_manager_disk_size(self.template))
        self.add_to_parameters(parameters.add_parameter_manager_disk_type(self.template))

        self.add_to_parameters(parameters.add_parameter_enable_system_prune(self.template))
        self.add_to_parameters(parameters.add_parameter_enable_cloudwatch_logs(self.template))

    def add_outputs(self):
        outputs.add_output_managers(self.template)
        outputs.add_output_dns_target(self.template)
        outputs.add_output_az_warning(self.template)

    def dynamodb(self):
        # dynamodb table
        resources.add_resource_dyn_table(self.template)

    def worker_userdata_head(self):
        return resources.worker_node_userdata_head()

    def workder_userdata_body(self):
        return resources.worker_node_userdata_body()

    def worker_userdata(self):
        header = ["#!/bin/sh\n"]
        head_data = self.worker_userdata_head()
        body_data = self.workder_userdata_body()
        data = header + head_data + body_data
        return Base64(Join("", data))

    def manager_userdata_head(self):
        return resources.manager_node_userdata_head()

    def manager_userdata_body(self):
        return resources.manager_node_userdata_body()

    def manager_userdata(self):
        header = ["#!/bin/sh\n"]
        head_data = self.manager_userdata_head()
        body_data = self.manager_userdata_body()
        data = header + head_data + body_data
        return Base64(Join("", data))

    def add_resources(self):
        # Networking
        self.vpc()

        # logs
        self.logs()

        # load Balancer
        self.load_balancer()

        # IAM
        self.iam()

        # security groups
        self.security_groups()

        # autoscaling
        self.autoscaling()

        # dynamodb table
        self.dynamodb()

        # sqs queues
        self.sqs()

    def vpc(self):
        if self.create_vpc:
            resources.add_resource_vpc(self.template)
            resources.add_resource_subnet_az_1(self.template)
            resources.add_resource_subnet_az_2(self.template)
            resources.add_resource_subnet_az_3(self.template)
            resources.add_resource_internet_gateway(self.template)
            resources.add_resource_attach_gateway(self.template)
            resources.add_resource_route_table(self.template)
            resources.add_resource_public_route(self.template)
            resources.add_resource_subnet1_route_table(self.template)
            resources.add_resource_subnet2_route_table(self.template)
            resources.add_resource_subnet3_route_table(self.template)

    def logs(self):
        resources.add_resource_log_group(self.template)

    def load_balancer(self):
        resources.add_resource_external_lb(self.template, self.create_vpc)

    def iam(self):
        resources.add_resource_proxy_role(self.template)
        resources.add_resource_IAM_dyn_policy(self.template)
        resources.add_resource_iam_swarm_api_policy(self.template)
        resources.add_resource_iam_sqs_policy(self.template)
        resources.add_resource_iam_sqs_cleanup_policy(self.template)
        resources.add_resource_iam_autoscale_policy(self.template)
        resources.add_resource_iam_elb_policy(self.template)
        resources.add_resource_iam_instance_profile(self.template)

    def security_groups(self):
        # security groups
        resources.add_resource_swarm_wide_security_group(self.template, self.create_vpc)
        resources.add_resource_manager_security_group(self.template)
        resources.add_resource_worker_security_group(self.template, self.create_vpc)
        resources.add_resource_external_lb_sg(self.template, self.create_vpc)

    def sqs(self):
        # SQS
        resources.add_resources_sqs_cleanup(self.template)
        resources.add_resources_sqs_swarm(self.template)

    def autoscaling(self):
        # scaling groups
        resources.add_resource_manager_upgrade_hook(self.template)
        resources.add_resource_worker_upgrade_hook(self.template)

        # manager
        manager_launch_config_name = u'ManagerLaunchConfig{}'.format(self.flat_edition_version_upper)
        resources.add_resource_manager_autoscalegroup(
            self.template, self.create_vpc, manager_launch_config_name)
        resources.add_resource_manager_launch_config(self.template, self.manager_userdata(),
                                                     launch_config_name=manager_launch_config_name)
        # worker
        worker_launch_config_name = u'NodeLaunchConfig{}'.format(self.flat_edition_version_upper)
        resources.add_resource_worker_autoscalegroup(self.template, worker_launch_config_name)
        resources.add_resource_worker_launch_config(self.template, self.worker_userdata(),
                                                    launch_config_name=worker_launch_config_name)

    def generate_template(self):
        return self.template.to_json()

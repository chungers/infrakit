from troposphere import Template, Base64, Join

from sections import mappings
from sections import parameters
from sections import metadata
from sections import conditions
from sections import outputs
from sections import resources


class AWSBaseTemplate(object):

    def __init__(self, docker_version,
                 docker_for_aws_version, edition_addon, channel, amis,
                 create_vpc=True, template_description=None,
                 use_ssh_cidr=False,
                 experimental_flag=True,
                 has_ddc=False):
        self.template = Template()
        self.parameters = {}
        self.parameter_labels = {}
        self.docker_version = docker_version
        self.edition_addon = edition_addon
        self.channel = channel
        self.amis = amis
        self.has_ddc = has_ddc
        self.create_vpc = create_vpc
        self.template_description = template_description
        self.use_ssh_cidr = use_ssh_cidr
        self.experimental_flag = experimental_flag
        self.docker_for_aws_version = docker_for_aws_version

        flat_edition_version = docker_for_aws_version.replace(
            " ", "").replace("_", "").replace("-", "").replace(".", "")
        self.flat_edition_version = flat_edition_version
        flat_edition_version_upper = flat_edition_version.capitalize()
        self.flat_edition_version_upper = flat_edition_version_upper

    def build(self):
        self.add_template_version()
        self.add_template_description()
        self.add_parameters()
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
            description = u"Docker for AWS {}".format(
                self.docker_for_aws_version)
        self.template.add_description(description)

    def get_parameter_groups(self):
        return metadata.parameter_groups()

    def parameter_groups(self):
        """ Override to add more parameters """
        return self.get_parameter_groups()

    def add_metadata(self):
        metadata.metadata(self.template,
                          self.parameter_groups(),
                          self.parameter_labels)

    def add_conditions(self):
        conditions.add_condition_create_log_resources(self.template)
        conditions.add_condition_hasonly2AZs(self.template)
        conditions.add_condition_EBSOptimized(self.template)

    def add_mapping_version(self):
        mappings.add_mapping_version(
            self.template, self.docker_version,
            self.docker_for_aws_version, self.edition_addon,
            self.channel, self.has_ddc)

    def add_mappings(self):
        self.add_aws2az_mapping()
        self.add_mapping_version()
        mappings.add_mapping_vpc_cidrs(self.template)
        mappings.add_mapping_instance_type_2_arch(self.template)
        mappings.add_mapping_amis(self.template, self.amis)

    def add_aws2az_mapping(self):
        """ Add the aws2az mapping to the template.
        Override this method to change"""
        self.template.add_mapping('AWSRegion2AZ', mappings.aws2az_data())

    def add_to_parameters(self, result):
        key, value = result
        self.parameter_labels[key] = value

    def add_parameter_instancetype(self):
        self.add_to_parameters(
            parameters.add_parameter_instancetype(
                self.template))

    def add_parameter_manager_instancetype(self):
        self.add_to_parameters(
            parameters.add_parameter_manager_instancetype(
                self.template))

    def add_parameter_manager_disk_type(self):
        self.add_to_parameters(
            parameters.add_parameter_manager_disk_type(self.template))

    def add_parameter_manager_cluster_size(self):
        self.add_to_parameters(
            parameters.add_parameter_manager_size(self.template))

    def add_parameter_cloudwatch_logs(self):
        self.add_to_parameters(
            parameters.add_parameter_enable_cloudwatch_logs(self.template))

    def add_parameters(self):
        self.add_to_parameters(parameters.add_parameter_keyname(self.template))

        # instance typees
        self.add_parameter_instancetype()
        self.add_parameter_manager_instancetype()

        self.add_to_parameters(
            parameters.add_parameter_cluster_size(self.template))
        self.add_to_parameters(
            parameters.add_parameter_worker_disk_size(self.template))
        self.add_to_parameters(
            parameters.add_parameter_worker_disk_type(self.template))

        self.add_parameter_manager_cluster_size()

        self.add_to_parameters(
            parameters.add_parameter_manager_disk_size(self.template))
        self.add_parameter_manager_disk_type()

        self.add_to_parameters(
            parameters.add_parameter_enable_system_prune(self.template))
        self.add_parameter_cloudwatch_logs()

        self.add_to_parameters(
            parameters.add_parameter_enable_ebs_optimized(self.template))

        self.add_to_parameters(
            parameters.add_parameter_enable_cloudstor_efs(self.template))


    def add_outputs(self):
        outputs.add_output_managers(self.template)
        outputs.add_output_dns_target(self.template)
        outputs.add_output_az_warning(self.template)
        outputs.add_output_elb_zone_id(self.template)
        outputs.add_output_security_groups(self.template)

    def dynamodb(self):
        # dynamodb table
        resources.add_resource_dyn_table(self.template)

    def worker_userdata_head(self, instance_name=None):
        return resources.worker_node_userdata_head(
            experimental_flag=self.experimental_flag,
            instance_name=instance_name)

    def worker_userdata_body(self):
        return resources.worker_node_userdata_body()

    def userdata_header(self):
        return ["#!/bin/sh\n", ]
        # commenting out since it makes debugging issues harder
        # add it back in once things are stable again.
        # "# Close STDOUT file descriptor\n",
        # "exec 1<&-\n",
        # "# Close STDERR FD\n",
        # "exec 2<&-\n",
        # "# Open STDOUT as log file for read and write.\n",
        # "exec 1<>/var/lib/docker/editions.log\n",
        # "# Redirect STDERR to STDOUT\n",
        # "exec 2>&1\n"]

    def worker_userdata(self):
        header = self.userdata_header()
        head_data = self.worker_userdata_head()
        body_data = self.worker_userdata_body()
        data = header + head_data + body_data
        return Base64(Join("", data))

    def manager_userdata_head(self):
        return resources.manager_node_userdata_head(
            experimental_flag=self.experimental_flag)

    def manager_userdata_body(self):
        return resources.manager_node_userdata_body()

    def manager_userdata(self):
        header = self.userdata_header()
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

        # Custom resources
        self.custom()

        # lambda functions
        self.awslambda()

        # s3
        self.s3()

        # EFS
        self.efs()

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
            outputs.add_output_vpcid(self.template)

    def s3(self):
        pass

    def logs(self):
        resources.add_resource_log_group(self.template)

    def load_balancer(self):
        resources.add_resource_external_lb(self.template, self.create_vpc)

    def iam_dyn(self):
        # overridable for DTR role
        resources.add_resource_iam_dyn_policy(self.template)

    def iam_sqs(self):
        # overridable for DTR role
        resources.add_resource_iam_sqs_policy(self.template)
        resources.add_resource_iam_sqs_cleanup_policy(self.template)

    def iam_autoscale(self):
        # overridable for DTR role
        resources.add_resource_iam_autoscale_policy(self.template)

    def iam_log(self):
        # overridable for DTR role
        resources.add_resource_iam_log_policy(self.template)

    def iam(self):
        resources.add_resource_proxy_role(self.template)

        resources.add_resource_iam_swarm_api_policy(self.template)
        resources.add_resource_iam_elb_policy(self.template)
        resources.add_resource_iam_instance_profile(self.template)
        resources.add_resource_iam_cloudstor_ebs_policy(self.template)

        self.iam_dyn()
        self.iam_sqs()
        self.iam_autoscale()

        # worker
        resources.add_resource_worker_iam_role(self.template)
        resources.add_resource_iam_worker_dyn_policy(self.template)
        resources.add_resource_iam_worker_instance_profile(self.template)
        self.iam_log()

    def security_groups(self):
        # security groups
        resources.add_resource_swarm_wide_security_group(
            self.template, self.create_vpc)
        resources.add_resource_manager_security_group(
            self.template, use_ssh_cidr=self.use_ssh_cidr)
        resources.add_resource_worker_security_group(
            self.template, self.create_vpc)
        resources.add_resource_external_lb_sg(
            self.template, self.create_vpc)

    def sqs(self):
        # SQS
        resources.add_resources_sqs_cleanup(self.template)
        resources.add_resources_sqs_swarm(self.template)

    def autoscaling_managers(self, manager_launch_config_name):
        lb_list = ["ExternalLoadBalancer", ]
        resources.add_resource_manager_autoscalegroup(
            self.template, self.create_vpc, manager_launch_config_name,
            lb_list)
    
    def autoscaling_workers(self, worker_launch_config_name):
        resources.add_resource_worker_autoscalegroup(
            self.template, worker_launch_config_name)

    def autoscaling(self):
        # scaling groups
        resources.add_resource_manager_upgrade_hook(self.template)
        resources.add_resource_worker_upgrade_hook(self.template)

        # manager
        manager_launch_config_name = u'ManagerLaunchConfig{}'.format(
            self.flat_edition_version_upper)
        self.autoscaling_managers(manager_launch_config_name)
        resources.add_resource_manager_launch_config(
            self.template, self.manager_userdata(),
            launch_config_name=manager_launch_config_name)
        # worker
        worker_launch_config_name = u'NodeLaunchConfig{}'.format(
            self.flat_edition_version_upper)
        self.autoscaling_workers(worker_launch_config_name)
        resources.add_resource_worker_launch_config(
            self.template, self.worker_userdata(),
            launch_config_name=worker_launch_config_name)

    def custom(self):
        # custom resources
        if self.create_vpc:
            resources.add_resource_custom_az_info(self.template)

    def awslambda(self):
        if self.create_vpc:
            resources.add_resource_az_info_function(self.template)
            resources.add_resource_iam_lambda_execution_role(self.template)
            conditions.add_condition_LambdaSupported(self.template)

    def efs(self):
        # efs
        resources.add_resource_efs(self.template)
        resources.add_resource_mount_targets(self.template)
        conditions.add_condition_EFSSupported(self.template)

    def generate_template(self):
        return self.template.to_json()

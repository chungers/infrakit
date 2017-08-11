from os import path
from os import getenv
from troposphere import (
    Parameter, Ref, Output, GetAtt, Join, FindInMap, Base64)

from docker_ee import DockerEEVPCTemplate, DockerEEVPCExistingTemplate
from sections import mappings
from sections import resources
from sections import parameters
from sections import constants

DTR_TAG = '2.2.7'
UCP_TAG = '2.1.5'
UCP_INIT_TAG = getenv('UCP_INIT_TAG', '17.06.1-ce-aws1')
DTR_INIT_TAG = getenv('DTR_INIT_TAG', '17.06.1-ce-aws1')


class DDCVPCTemplate(DockerEEVPCTemplate):

    def __init__(self, docker_version, docker_for_aws_version,
                 edition_addon, channel, amis,
                 create_vpc=True, template_description=None,
                 use_ssh_cidr=True,
                 experimental_flag=False,
                 has_ddc=True):
        if not template_description:
            template_description = u"Docker EE DDC for AWS {}".format(
                docker_for_aws_version)
        super(DDCVPCTemplate, self).__init__(
            docker_version, docker_for_aws_version,
            edition_addon, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag,
            has_ddc=has_ddc)

    def parameter_groups(self):
        parameter_groups = super(DDCVPCTemplate, self).parameter_groups()
        parameter_groups.append(
            {"Label": {"default": "DDC Properties"},
             "Parameters": ["DDCUsernameSet", "DDCPasswordSet",
                            "License"]}
        )
        return parameter_groups

    def add_parameter_instancetype(self):
        self.add_to_parameters(
            parameters.add_parameter_instancetype(
                self.template,
                default_instance_type='t2.large',
                instance_types=constants.DDC_WORKER_INSTANCE_TYPES))

    def add_parameter_manager_instancetype(self):
        self.add_to_parameters(
            parameters.add_parameter_manager_instancetype(
                self.template,
                default_instance_type='m4.large',
                instance_types=constants.DDC_INSTANCE_TYPES))
    
    def add_parameter_manager_disk_type(self):
        self.add_to_parameters(
            parameters.add_parameter_manager_disk_type(
                self.template,
                default_disk_type='gp2'))

    def add_parameter_manager_cluster_size(self):
        """ don't let them put only 1 swarm manager """
        allowed_values = ["3", "5"]
        self.add_to_parameters(
            parameters.add_parameter_manager_size(
                self.template, allowed_values=allowed_values))

    def add_parameter_cloudwatch_logs(self):
        """ Cloudwatch logging doesn't work well with UCP, set default to no"""
        self.add_to_parameters(
            parameters.add_parameter_enable_cloudwatch_logs(
                self.template, default='no'))

    def add_parameters(self):
        super(DDCVPCTemplate, self).add_parameters()
        self.add_ddc_license()
        self.add_ddc_username()
        self.add_ddc_password()

    def add_ddc_license(self):
        self.template.add_parameter(Parameter(
            "License",
            Description="Docker Datacenter License in JSON format or URL "
                        "to download it. Get Trial License here "
                        "https://store.docker.com/bundles/docker-datacenter",
            Type='String',
            NoEcho=True
        ))
        self.add_to_parameters((
            'License',
            {"default": "Enter your Docker Datacenter License"}))

    def add_ddc_username(self):
        self.template.add_parameter(Parameter(
            "DDCUsernameSet",
            ConstraintDescription="Please enter the username you want to use "
                                  "for Docker Datacenter",
            Description="Docker Datacenter Username?",
            Type='String',
            Default="admin",
        ))
        self.add_to_parameters((
            'DDCUsernameSet',
            {"default":
             "Enter the Username you want to use with Docker Datacenter"}))

    def add_ddc_password(self):
        self.template.add_parameter(Parameter(
            "DDCPasswordSet",
            ConstraintDescription="must be at least 8 characters",
            Description="Docker Datacenter Password?",
            MaxLength="40",
            MinLength="8",
            Type='String',
            NoEcho=True
        ))
        self.add_to_parameters((
            'DDCPasswordSet',
            {"default": "Enter your Docker Datacenter password"}))

    def add_outputs(self):
        """ add outputs for DDC
        """
        super(DDCVPCTemplate, self).add_outputs()
        self.template.add_output(Output(
            "DDCUsername",
            Description="Docker Datacenter Username",
            Value=Ref("DDCUsernameSet"))
        )
        self.template.add_output(Output(
            "DTRLoginURL",
            Description="Docker Datacenter DTR Login URL",
            Value=Join("", ["https://",
                            GetAtt("DTRLoadBalancer", "DNSName")])
        ))
        self.template.add_output(Output(
            "UCPLoginURL",
            Description="Docker Datacenter UCP Login URL",
            Value=Join("", ["https://",
                            GetAtt("UCPLoadBalancer", "DNSName")])
        ))

    def s3(self):
        super(DDCVPCTemplate, self).s3()
        resources.add_s3_dtr_bucket(self.template)

    def iam(self):
        super(DDCVPCTemplate, self).iam()
        resources.add_resource_s3_ddc_bucket_policy(self.template)

    def autoscaling_managers(self, manager_launch_config_name):
        """ Overrides the base method, to include the two DDC ELBs"""
        lb_list = ["ExternalLoadBalancer", "UCPLoadBalancer", "DTRLoadBalancer"]
        resources.add_resource_manager_autoscalegroup(
            self.template, self.create_vpc, manager_launch_config_name,
            lb_list, health_check_grace_period=1200, timeout='PT1H')

    def autoscaling_workers(self, worker_launch_config_name):
        """ Overrides the base method, to include the create timeout"""
        resources.add_resource_worker_autoscalegroup(
            self.template, worker_launch_config_name, timeout='PT1H')

    def load_balancer(self):
        super(DDCVPCTemplate, self).load_balancer()
        resources.add_resource_ddc_ucp_lb(self.template, self.create_vpc)
        resources.add_resource_ddc_dtr_lb(self.template, self.create_vpc)

    def security_groups(self):
        # security groups
        super(DDCVPCTemplate, self).security_groups()
        resources.add_resource_ddc_ucp_lb_sg(self.template, self.create_vpc)
        resources.add_resource_ddc_dtr_lb_sg(self.template, self.create_vpc)

    def add_mapping_version(self):
        extra_data = {
            'DTRTAG': DTR_TAG,
            'UCPTAG': UCP_TAG,
            'UCPINITTAG': UCP_INIT_TAG,
            'DTRINITTAG': DTR_INIT_TAG,
        }
        mappings.add_mapping_version(
            self.template, self.docker_version,
            self.docker_for_aws_version, self.edition_addon,
            self.channel, self.has_ddc, extra_data=extra_data)

    def common_userdata_head(self):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        orig_data = super(DDCVPCTemplate, self).common_userdata_head()
        data = [
            "export UCP_ADMIN_USER='", Ref("DDCUsernameSet"), "'\n",
            "export UCP_ADMIN_PASSWORD='", Ref("DDCPasswordSet"), "'\n",
            "export S3_BUCKET_NAME='", Ref("DDCBucket"), "'\n",
            "export LICENSE='", Ref("License"), "'\n",
            "export UCP_ELB_HOSTNAME='", GetAtt("UCPLoadBalancer", "DNSName"), "'\n",
            "export DTR_ELB_HOSTNAME='", GetAtt("DTRLoadBalancer", "DNSName"), "'\n",
            "export APP_ELB_HOSTNAME='", GetAtt("ExternalLoadBalancer", "DNSName"), "'\n",
            "export MANAGER_COUNT='", Ref("ManagerSize"), "'\n",
            "export UCP_TAG='", FindInMap("DockerForAWS", "version", "UCPTAG"), "'\n",
            "export DTR_TAG='", FindInMap("DockerForAWS", "version", "DTRTAG"), "'\n",
            "export UCP_INIT_TAG='", FindInMap("DockerForAWS", "version", "UCPINITTAG"), "'\n",
            "export DTR_INIT_TAG='", FindInMap("DockerForAWS", "version", "DTRINITTAG"), "'\n",
        ]
        return orig_data + data

    def common_userdata_body(self):
        """ This is the body of the userdata """
        script_dir = path.dirname(__file__)
        ddc_path = path.relpath("data/ddc/common_userdata.sh")
        ddc_file_path = path.join(script_dir, ddc_path)
        data = resources.userdata_from_file(ddc_file_path)
        return data

    def manager_userdata_body(self):
        """ This is the body of the userdata """
        orig_data = super(DDCVPCTemplate, self).manager_userdata_body()
        data = self.common_userdata_body()
        return orig_data + data

    def worker_userdata_body(self):
        """ This is the body of the userdata """
        orig_data = super(DDCVPCTemplate, self).worker_userdata_body()
        data = self.common_userdata_body()
        return orig_data + data


class DDCVPCExistingTemplate(DDCVPCTemplate, DockerEEVPCExistingTemplate):
    """ DDC Template for existing VPC."""
    def __init__(self, docker_version,
                 docker_for_aws_version, edition_addon, channel, amis,
                 create_vpc=False, template_description=None, has_ddc=True,
                 experimental_flag=False,
                 use_ssh_cidr=True):
        super(DDCVPCExistingTemplate, self).__init__(
            docker_version, docker_for_aws_version, edition_addon,
            channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            has_ddc=has_ddc,
            experimental_flag=experimental_flag,
            use_ssh_cidr=use_ssh_cidr
            )

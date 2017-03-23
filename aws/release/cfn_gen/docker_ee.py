from os import path

from troposphere import Parameter, Ref

from base import AWSBaseTemplate
from sections import mappings
from sections import resources
from existing_vpc import ExistingVPCTemplate


class DockerEEVPCTemplate(AWSBaseTemplate):

    def __init__(self, docker_version, docker_for_aws_version, 
                 edition_addon, channel, amis,
                 create_vpc=True, template_description=None,
                 use_ssh_cidr=True,
                 experimental_flag=False,
                 has_ddc=False):
        if not template_description:
            template_description = u"Docker EE for AWS {} ({})".format(
                docker_version, docker_for_aws_version)
        super(DockerEEVPCTemplate, self).__init__(
            docker_version, docker_for_aws_version, 
            edition_addon, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag,
            has_ddc=has_ddc)

    def add_parameters(self,
                       manager_default_instance_type=None,
                       worker_default_instance_type=None):
        super(DockerEEVPCTemplate, self).add_parameters(
            manager_default_instance_type=manager_default_instance_type,
            worker_default_instance_type=worker_default_instance_type
            )
        self.add_parameter_ssh_cidr()
        self.add_parameter_http_proxy()
        self.add_parameter_https_proxy()
        self.add_parameter_no_proxy()

    def parameter_groups(self):
        """ We need to add RemoteSSH to the Swarm properites
        parameter group after KeyName """
        parameter_groups = super(DockerEEVPCTemplate, self).parameter_groups()
        new_groups = []
        for group in parameter_groups:
            if group.get('Label', {}).get('default') == "Swarm Properties":
                params = group.get('Parameters', [])
                new_params = []
                for param in params:
                    new_params.append(param)
                    if param == "KeyName":
                        new_params.append("RemoteSSH")
                new_groups.append({
                    "Label": {"default": "Swarm Properties"},
                    "Parameters": new_params
                    })
            else:
                new_groups.append(group)

        # Add HTTP proxy
        new_groups.append(
            {"Label": {"default": "HTTP Proxy"},
             "Parameters": ["HTTPProxy", "HTTPSProxy", "NoProxy", ]}
        )
        return new_groups

    def add_aws2az_mapping(self):
        """ TODO: Remove this method when when EFS goes live on EE
        We are overriding to disable in EE templates."""
        data = mappings.aws2az_data()
        data['eu-west-1']['EFSSupport'] = 'no'
        data['us-east-1']['EFSSupport'] = 'no'
        data['us-east-2']['EFSSupport'] = 'no'
        data['us-west-2']['EFSSupport'] = 'no'

        self.template.add_mapping('AWSRegion2AZ', data)

    def add_parameter_ssh_cidr(self):
        self.template.add_parameter(Parameter(
            'RemoteSSH',
            Type='String',
            MaxLength="18",
            MinLength="9",
            ConstraintDescription="Must be a valid IP CIDR range of the form x.x.x.x/x.",
            Description="The IP address range that can SSH to the EC2 instance."))
        self.add_to_parameters(
            ('RemoteSSH',
             {"default": "Which IPs are allowed to SSH? [0.0.0.0/0 will allow SSH from anywhere]"}))

    def add_parameter_http_proxy(self):
        self.template.add_parameter(Parameter(
            'HTTPProxy',
            Type='String',
            AllowedPattern="^\S*$",
            ConstraintDescription="HTTP_PROXY environment variable setting",
            Description="Value for HTTP_PROXY environment variable."))
        self.add_to_parameters(
            ('HTTPProxy',
             {"default": "Value for HTTP_PROXY environment variable."}))

    def add_parameter_https_proxy(self):
        self.template.add_parameter(Parameter(
            'HTTPSProxy',
            Type='String',
            AllowedPattern="^\S*$",
            ConstraintDescription="HTTPS_PROXY environment variable setting",
            Description="Value for HTTPS_PROXY environment variable."))
        self.add_to_parameters(
            ('HTTPSProxy',
             {"default": "Value for HTTPS_PROXY environment variable."}))

    def add_parameter_no_proxy(self):
        self.template.add_parameter(Parameter(
            'NoProxy',
            Type='String',
            AllowedPattern="^\S*$",
            ConstraintDescription="NO_PROXY environment variable setting",
            Description="Value for NO_PROXY environment variable."))
        self.add_to_parameters(
            ('NoProxy',
             {"default": "Value for NO_PROXY environment variable."}))

    def common_userdata_head(self):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        data = [
            "export HTTP_PROXY='", Ref("HTTPProxy"), "'\n",
            "export HTTPS_PROXY='", Ref("HTTPSProxy"), "'\n",
            "export NO_PROXY='", Ref("NoProxy"), "'\n",
        ]
        return data

    def manager_userdata_head(self):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        orig_data = super(DockerEEVPCTemplate, self).manager_userdata_head()
        data = self.common_userdata_head()
        return orig_data + data

    def worker_userdata_head(self):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        orig_data = super(DockerEEVPCTemplate, self).worker_userdata_head()
        data = self.common_userdata_head()
        return orig_data + data

    def common_userdata_body(self):
        """ This is the body of the userdata """
        script_dir = path.dirname(__file__)

        ddc_path = path.relpath("data/ee/common_userdata.sh")
        ddc_file_path = path.join(script_dir, ddc_path)
        data = resources.userdata_from_file(ddc_file_path)
        return data

    def manager_userdata_body(self):
        """ This is the body of the userdata """
        orig_data = super(DockerEEVPCTemplate, self).manager_userdata_body()
        data = self.common_userdata_body()
        # we want our data to go above the original data.
        # we have some settings that need to be set before we restart docker.
        return data + orig_data

    def worker_userdata_body(self):
        """ This is the body of the userdata """
        orig_data = super(DockerEEVPCTemplate, self).worker_userdata_body()
        data = self.common_userdata_body()
        # we want our data to go above the original data.
        # we have some settings that need to be set before we restart docker.
        return data + orig_data


class DockerEEVPCExistingTemplate(DockerEEVPCTemplate, ExistingVPCTemplate):
    """ Cloud Template for existing VPC."""
    def __init__(self, docker_version, docker_for_aws_version, 
                 edition_addon, channel, amis,
                 create_vpc=False, template_description=None,
                 use_ssh_cidr=True,
                 experimental_flag=False,
                 has_ddc=False):
        super(DockerEEVPCExistingTemplate, self).__init__(
            docker_version, docker_for_aws_version, 
            edition_addon, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag,
            has_ddc=has_ddc
            )

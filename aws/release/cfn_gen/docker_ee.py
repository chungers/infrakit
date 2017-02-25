from troposphere import Parameter

from base import AWSBaseTemplate
from sections import mappings
from existing_vpc import ExistingVPCTemplate


class DockerEEVPCTemplate(AWSBaseTemplate):

    def __init__(self, docker_version, edition_version,
                 docker_for_aws_version, channel, amis,
                 create_vpc=True, template_description=None,
                 use_ssh_cidr=True,
                 experimental_flag=False):
        if not template_description:
            template_description = u"Docker EE for AWS {} ({})".format(
                docker_version, edition_version)
        super(DockerEEVPCTemplate, self).__init__(
            docker_version, edition_version,
            docker_for_aws_version, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag)

    def add_paramaters(self):
        super(DockerEEVPCTemplate, self).add_paramaters()
        self.add_parameter_ssh_cidr()

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


class DockerEEVPCExistingTemplate(DockerEEVPCTemplate, ExistingVPCTemplate):
    """ Cloud Template for existing VPC."""
    def __init__(self, docker_version, edition_version,
                 docker_for_aws_version, channel, amis,
                 create_vpc=False, template_description=None,
                 use_ssh_cidr=True,
                 experimental_flag=False):
        super(DockerEEVPCExistingTemplate, self).__init__(
            docker_version, edition_version,
            docker_for_aws_version,
            channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag
            )

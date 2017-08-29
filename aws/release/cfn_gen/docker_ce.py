from base import AWSBaseTemplate
from sections import parameters, conditions, resources
from existing_vpc import ExistingVPCTemplate
from os import path
from troposphere import If, Ref, FindInMap


class DockerCEVPCTemplate(AWSBaseTemplate):
    """ Docker CE specific items, or items that don't belong in base yet,
    since it isn't stable enough for EE yet.
    """

    def __init__(self, docker_version, docker_for_aws_version,
                 edition_addon, channel, amis,
                 create_vpc=True, template_description=None,
                 use_ssh_cidr=False,
                 experimental_flag=True,
                 has_ddc=False):
        if not template_description:
            template_description = u"Docker CE for AWS {} ({})".format(
                docker_version, docker_for_aws_version)
        super(DockerCEVPCTemplate, self).__init__(
            docker_version, docker_for_aws_version,
            edition_addon, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag,
            has_ddc=has_ddc)


class DockerCEVPCExistingTemplate(DockerCEVPCTemplate, ExistingVPCTemplate):
    """ CE Template for existing VPC."""
    def __init__(self, docker_version, docker_for_aws_version,
                 edition_addon, channel, amis,
                 create_vpc=False, template_description=None,
                 use_ssh_cidr=False,
                 experimental_flag=True,
                 has_ddc=False):
        super(DockerCEVPCExistingTemplate, self).__init__(
            docker_version, docker_for_aws_version,
            edition_addon, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag,
            has_ddc=has_ddc
            )

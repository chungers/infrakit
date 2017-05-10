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

    def add_parameters(self):
        super(DockerCEVPCTemplate, self).add_parameters()
        self.add_to_parameters(
            parameters.add_parameter_enable_cloudstor(
                self.template))

    def parameter_groups(self):
        """ We need to add RemoteSSH to the Swarm properites
        parameter group after KeyName """
        parameter_groups = super(DockerCEVPCTemplate, self).parameter_groups()
        new_groups = []
        for group in parameter_groups:
            if group.get('Label', {}).get('default') == "Swarm Properties":
                params = group.get('Parameters', [])
                new_params = []
                for param in params:
                    new_params.append(param)
                    if param == "EnableCloudWatchLogs":
                        new_params.append("EnableCloudStor")
                new_groups.append({
                    "Label": {"default": "Swarm Properties"},
                    "Parameters": new_params
                    })
            else:
                new_groups.append(group)
        return new_groups

    def add_resources(self):
        super(DockerCEVPCTemplate, self).add_resources()

        # EFS
        self.efs()

    def efs(self):
        # efs
        resources.add_resource_efs(self.template)
        resources.add_resource_mount_targets(self.template)
        conditions.add_condition_EFSSupported(self.template)
        conditions.add_condition_CloudStor_selected(self.template)
        conditions.add_condition_InstallCloudStor(self.template)

    def common_userdata_head(self):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        data = [
            "export ENABLE_EFS='",
            If("InstallCloudStor", 'yes', 'no'), "'\n",

            "export EFS_ID_REGULAR='",
            If("InstallCloudStor", Ref("FileSystemGP"), ''), "'\n",

            "export EFS_ID_MAXIO='",
            If("InstallCloudStor", Ref("FileSystemMaxIO"), ''), "'\n",
        ]
        return data

    def manager_userdata_head(self):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        orig_data = super(DockerCEVPCTemplate, self).manager_userdata_head()
        data = self.common_userdata_head()
        return orig_data + data

    def worker_userdata_head(self, instance_name=None):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        orig_data = super(DockerCEVPCTemplate, self).worker_userdata_head(
            instance_name=instance_name)
        data = self.common_userdata_head()
        return orig_data + data

    def common_userdata_body(self):
        """ This is the body of the userdata """
        script_dir = path.dirname(__file__)

        ddc_path = path.relpath("data/ce/common_userdata.sh")
        ddc_file_path = path.join(script_dir, ddc_path)
        data = resources.userdata_from_file(ddc_file_path)
        return data

    def worker_userdata_body(self):
        orig_data = super(DockerCEVPCTemplate, self).worker_userdata_body()
        data = self.common_userdata_body()
        return orig_data + data

    def manager_userdata_body(self):
        orig_data = super(DockerCEVPCTemplate, self).manager_userdata_body()
        data = self.common_userdata_body()
        return orig_data + data


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

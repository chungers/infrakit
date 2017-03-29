from os import path
from troposphere import Parameter, Ref, Output

from ddc import DDCVPCTemplate, DDCVPCExistingTemplate, DTR_TAG, UCP_TAG, UCP_INIT_TAG
from sections import resources


class DDCDevVPCTemplate(DDCVPCTemplate):
    def __init__(self, docker_version,
                 docker_for_aws_version, edition_addon, channel, amis,
                 create_vpc=True, template_description=None,
                 use_ssh_cidr=False,
                 experimental_flag=False,
                 has_ddc=True):
        if not template_description:
            template_description = u"Docker EE DDC (DEV) for AWS {}".format(
                docker_version)
        super(DDCDevVPCTemplate, self).__init__(
            docker_version,
            docker_for_aws_version, edition_addon, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag,
            has_ddc=has_ddc)

    def common_userdata_head(self):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        orig_data = super(DDCDevVPCTemplate, self).common_userdata_head()
        data = [
            "export HUB_USERNAME='", Ref("HubUsername"), "'\n",
            "export HUB_PASSWORD='", Ref("HubPassword"), "'\n",
            "export UCP_NAMESPACE='", Ref("UCPNamespace"), "'\n",
            "export UCP_IMAGE_TAG='", Ref("UCPImageTag"), "'\n",
            "export DTR_NAMESPACE='", Ref("DTRNamespace"), "'\n",
            "export DTR_IMAGE_TAG='", Ref("DTRImageTag"), "'\n",
            "export UCP_INIT_TAG='", Ref("UCPInitTag"), "'\n",
        ]
        return orig_data + data

    def common_userdata_body_top(self):
        """ This is the body of the userdata """
        orig_data = super(DDCDevVPCTemplate, self).common_userdata_body_top()
        script_dir = path.dirname(__file__)

        ddc_path = path.relpath("data/ddc_dev/common_userdata_top.sh")
        ddc_file_path = path.join(script_dir, ddc_path)
        data = resources.userdata_from_file(ddc_file_path)
        return orig_data + data

    def common_userdata_body(self):
        """ This is the body of the userdata """
        script_dir = path.dirname(__file__)
        ddc_path = path.relpath("data/ddc_dev/common_userdata.sh")
        ddc_file_path = path.join(script_dir, ddc_path)
        data = resources.userdata_from_file(ddc_file_path)
        return data

    def parameter_groups(self):
        parameter_groups = super(DDCDevVPCTemplate, self).parameter_groups()
        parameter_groups.append(
            {"Label": {"default": "DDC Dev Properties"},
             "Parameters": ["HubUsername", "HubPassword", "UCPNamespace",
                            "UCPImageTag", "DTRNamespace", "DTRImageTag",
                            "UCPInitTag"]}
        )
        return parameter_groups

    def add_parameters(self):
        super(DDCDevVPCTemplate, self).add_parameters()
        self.add_hub_username()
        self.add_hub_password()
        self.add_ucp_namespace()
        self.add_ucp_imagetag()
        self.add_dtr_namespace()
        self.add_dtr_image_tag()
        self.add_ucp_init_tag()

    def add_hub_username(self):
        """
        "HubUsername": {
            "ConstraintDescription": "A Docker Hub account that has access to the specified Datacenter image on Docker Hub",
            "Description": "Docker Hub account for pulling Datacenter image?",
            "Type": "String"
        },
        """
        self.template.add_parameter(Parameter(
            "HubUsername",
            ConstraintDescription="A Docker Hub account that has access to the specified Datacenter image on Docker Hub",
            Description="Docker Hub account for pulling Datacenter image?",
            Type='String',
        ))
        self.add_to_parameters((
            'HubUsername',
            {"default": "Enter your Docker Hub username"}))

    def add_hub_password(self):
        """
        "HubPassword": {
            "ConstraintDescription": "Docker password corresponding to the Docker ID",
            "Description": "Docker password pulling image?",
            "NoEcho": "true",
            "Type": "String"
        },
        """
        self.template.add_parameter(Parameter(
            "HubPassword",
            ConstraintDescription="Docker password corresponding to the Docker ID",
            Description="Docker password pulling image?",
            NoEcho=True,
            Type='String',
        ))
        self.add_to_parameters((
            'HubPassword',
            {"default": "Enter your Docker Hub password"}))

    def add_ucp_namespace(self):
        """
         "UCPNamespace": {
            "AllowedValues": [
                "docker",
                "dockerorcadev"
            ],
            "Default": "dockerorcadev",
            "Description": "Hub namespace to pull Datacenter?",
            "Type": "String"
        },
        """
        self.template.add_parameter(Parameter(
            "UCPNamespace",
            AllowedValues=[
                "docker",
                "dockerorcadev"
            ],
            Default="dockerorcadev",
            Description="Hub namespace to pull Datacenter?",
            Type='String',
        ))
        self.add_to_parameters((
            'UCPNamespace',
            {"default": "Hub namespace to pull Datacenter?"}))

    def add_ucp_imagetag(self):
        """
           "UCPImageTag": {
                "ConstraintDescription": "Please enter the image tag you want to use for pulling UCP",
                "Default": "2.2.0-20ad8f6",
                "Description": "Hub tag to pull UCP?",
                "Type": "String"
            },
        """
        self.template.add_parameter(Parameter(
            "UCPImageTag",
            Default=UCP_TAG,
            ConstraintDescription="Please enter the image tag you want to use for pulling UCP",
            Description="Hub tag to pull UCP?",
            Type='String',
        ))
        self.add_to_parameters((
            'UCPImageTag',
            {"default": "Hub tag to pull UCP?"}))

    def add_ucp_init_tag(self):
        """
           "UCPImageTag": {
                "ConstraintDescription": "Please enter the image tag you want to use for pulling UCP",
                "Default": "2.2.0-20ad8f6",
                "Description": "Hub tag to pull UCP?",
                "Type": "String"
            },
        """
        self.template.add_parameter(Parameter(
            "UCPInitTag",
            Default=UCP_INIT_TAG,
            ConstraintDescription="Which Tag for docker4x/ucp-init-aws ?",
            Description="Which docker4x/ddc-init-aws Tag",
            Type='String',
        ))
        self.add_to_parameters((
            'UCPInitTag',
            {"default": "Which docker4x/ddc-init-aws Tag"}))

    def add_dtr_namespace(self):
        """ "DTRNamespace": {
            "AllowedValues": [
                "docker",
                "dockerhubenterprise"
            ],
            "Default": "dockerhubenterprise",
            "Description": "Hub namespace to pull DTR?",
            "Type": "String"
        },
        """
        self.template.add_parameter(Parameter(
            "DTRNamespace",
            AllowedValues=[
                "docker",
                "dockerhubenterprise"
            ],
            Default="dockerhubenterprise",
            Description="Hub namespace to pull DTR?",
            Type='String',
        ))
        self.add_to_parameters((
            'DTRNamespace',
            {"default": "Hub namespace to pull DTR?"}))

    def add_dtr_image_tag(self):
        """
        "DTRImageTag": {
            "ConstraintDescription": "Please enter the image tag you want to use for pulling DTR",
            "Default": "2.2.1",
            "Description": "Hub tag to pull DTR?",
            "Type": "String"
        },
        """
        self.template.add_parameter(Parameter(
            "DTRImageTag",
            ConstraintDescription="Please enter the image tag you want to use for pulling DTR",
            Default=DTR_TAG,
            Description="Hub tag to pull DTR?",
            Type='String',
        ))
        self.add_to_parameters((
            'DTRImageTag',
            {"default": "Hub Tag to pull DTR?"}))

    def add_outputs(self):
        """

          "TheHubUsername": {
                "Description": "Docker ID",
                "Value": {
                    "Ref": "HubUsername"
                }
            },
            "TheUCPNamespace": {
                "Description": "Docker Hub UCP namespace",
                "Value": {
                    "Ref": "UCPNamespace"
                }
            },
            "TheUCPTag": {
                "Description": "Docker Hub tag",
                "Value": {
                    "Ref": "UCPImageTag"
                }
            },
            "TheDTRNamespace": {
                "Description": "Docker Hub DTR namespace",
                "Value": {
                    "Ref": "DTRNamespace"
                }
            },
            "TheDTRTag": {
                "Description": "Docker Hub tag",
                "Value": {
                    "Ref": "DTRImageTag"
                }
            },
        """
        super(DDCDevVPCTemplate, self).add_outputs()
        self.template.add_output(Output(
            "TheHubUsername",
            Description="Docker ID",
            Value=Ref("HubUsername"))
        )
        self.template.add_output(Output(
            "TheUCPNamespace",
            Description="Docker Hub UCP namespace",
            Value=Ref("UCPNamespace"))
        )
        self.template.add_output(Output(
            "TheUCPTag",
            Description="Docker Hub UCP tag",
            Value=Ref("UCPImageTag"))
        )
        self.template.add_output(Output(
            "TheDTRNamespace",
            Description="Docker Hub DTR namespace",
            Value=Ref("DTRNamespace"))
        )
        self.template.add_output(Output(
            "TheDTRTag",
            Description="Docker Hub DTR tag",
            Value=Ref("DTRImageTag"))
        )


class DDCDevVPCExistingTemplate(DDCDevVPCTemplate, DDCVPCExistingTemplate):
    """ DDC Template for existing VPC."""
    def __init__(self, docker_version,
                 docker_for_aws_version, edition_addon, channel, amis,
                 create_vpc=False, template_description=None, has_ddc=True,
                 use_ssh_cidr=False):
        super(DDCDevVPCExistingTemplate, self).__init__(
            docker_version,
            docker_for_aws_version, edition_addon,
            channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            has_ddc=has_ddc,
            use_ssh_cidr=use_ssh_cidr
            )

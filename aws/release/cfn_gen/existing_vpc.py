from troposphere import Parameter

from base import AWSBaseTemplate
from sections import mappings


class ExistingVPCTemplate(AWSBaseTemplate):

    def __init__(self, docker_version, docker_for_aws_version, 
                 edition_addon, channel, amis,
                 create_vpc=False, template_description=None,
                 use_ssh_cidr=False,
                 experimental_flag=True,
                 has_ddc=False):
        super(ExistingVPCTemplate, self).__init__(
            docker_version, docker_for_aws_version, 
            edition_addon, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description,
            use_ssh_cidr=use_ssh_cidr,
            experimental_flag=experimental_flag,
            has_ddc=has_ddc
            )

    def add_parameters(self,
                       manager_default_instance_type=None,
                       worker_default_instance_type=None):
        super(ExistingVPCTemplate, self).add_parameters(
            manager_default_instance_type=manager_default_instance_type,
            worker_default_instance_type=worker_default_instance_type)
        self.add_vpc_param()
        self.add_subnet1()
        self.add_subnet2()
        self.add_subnet3()
        self.add_vpc_cidr()

    def parameter_groups(self):
        parameter_groups = super(ExistingVPCTemplate, self).parameter_groups()
        parameter_groups.append(
            {"Label": {"default": "VPC/Network"},
             "Parameters": ["Vpc", "VpcCidr",
                            "PubSubnetAz1", "PubSubnetAz2",
                            "PubSubnetAz3"]}
        )
        return parameter_groups

    def add_vpc_param(self):
        self.template.add_parameter(Parameter(
            "Vpc",
            Description="VPC must have internet access "
                        "(with Internet Gateway or Virtual Private Gateway)",
            Type='AWS::EC2::VPC::Id'
        ))
        self.add_to_parameters(('Vpc', {"default": "VPC"}))

    def add_subnet1(self):
        self.template.add_parameter(Parameter(
            "PubSubnetAz1",
            Description="Public Subnet 1",
            Type='AWS::EC2::Subnet::Id'
        ))
        self.add_to_parameters(
            ('PubSubnetAz1', {"default": "Public Subnet 1"}))

    def add_subnet2(self):
        self.template.add_parameter(Parameter(
            "PubSubnetAz2",
            Description="Public Subnet 2",
            Type='AWS::EC2::Subnet::Id'
        ))
        self.add_to_parameters(
            ('PubSubnetAz2', {"default": "Public Subnet 2"}))

    def add_subnet3(self):
        self.template.add_parameter(Parameter(
            "PubSubnetAz3",
            Description="Public Subnet 3",
            Type='AWS::EC2::Subnet::Id'
        ))
        self.add_to_parameters(
            ('PubSubnetAz3',
             {"default": "Public Subnet 3"}))

    def add_vpc_cidr(self):
        self.template.add_parameter(Parameter(
            "VpcCidr",
            Type='String',
            ConstraintDescription="Must be a valid IP CIDR range of the form x.x.x.x/x.",
            Description="The CIDR range for your VPC in form x.x.x.x/x",
        ))
        self.add_to_parameters(
            ('VpcCidr',
             {"default": "VPC CIDR Range"}))

    def add_aws2az_mapping(self):
        """ No need to have Lambda support when VPC is existing. """
        data = mappings.aws2az_data()
        for region in data.keys():
            data[region]['LambdaSupport'] = 'no'

        self.template.add_mapping('AWSRegion2AZ', data)

from troposphere import Parameter

from base import AWSBaseTemplate


class ExistingVPCTemplate(AWSBaseTemplate):

    def __init__(self, docker_version, edition_version,
                 docker_for_aws_version, channel, amis,
                 create_vpc=False, template_description=None):
        super(ExistingVPCTemplate, self).__init__(
            docker_version, edition_version,
            docker_for_aws_version,
            channel, amis,
            create_vpc=create_vpc,
            template_description=template_description
            )

    def add_paramaters(self):
        super(ExistingVPCTemplate, self).add_paramaters()
        self.add_vpc_param()
        self.add_subnet1()
        self.add_subnet2()
        self.add_subnet3()

    def parameter_groups(self):
        parameter_groups = super(ExistingVPCTemplate, self).parameter_groups()
        parameter_groups.append(
            {"Label": {"default": "VPC/Network"},
             "Parameters": ["Vpc", "PubSubnetAz1", "PubSubnetAz2", "PubSubnetAz3"]}
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
        self.add_to_parameters(('PubSubnetAz1', {"default": "Public Subnet 1"}))

    def add_subnet2(self):
        self.template.add_parameter(Parameter(
            "PubSubnetAz2",
            Description="Public Subnet 2",
            Type='AWS::EC2::Subnet::Id'
        ))
        self.add_to_parameters(('PubSubnetAz2', {"default": "Public Subnet 2"}))

    def add_subnet3(self):
        self.template.add_parameter(Parameter(
            "PubSubnetAz3",
            Description="Public Subnet 3",
            Type='AWS::EC2::Subnet::Id'
        ))
        self.add_to_parameters(('PubSubnetAz3', {"default": "Public Subnet 3"}))

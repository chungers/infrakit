from os import path

from troposphere import Parameter, And, Not, Equals, Ref, Output, Join, GetAtt

from base import AWSBaseTemplate

from sections import resources
from existing_vpc import ExistingVPCTemplate


class CloudVPCTemplate(AWSBaseTemplate):

    def __init__(self, docker_version, edition_version,
                 docker_for_aws_version, channel, amis,
                 create_vpc=True, template_description=None):
        if not template_description:
            template_description = u"Docker for AWS {} ({}) cloud".format(
                docker_version, edition_version)
        super(CloudVPCTemplate, self).__init__(
            docker_version, edition_version,
            docker_for_aws_version, channel, amis,
            create_vpc=create_vpc,
            template_description=template_description)

    def add_paramaters(self):
        super(CloudVPCTemplate, self).add_paramaters()
        self.add_cloud_clustername()
        self.add_cloud_username()
        self.add_cloud_apikey()
        self.add_cloud_rest_host()
        self.add_cloud_id_host()

    def manager_userdata_head(self):
        """ The Head of the userdata script, this is where
        you would declare all of your shell variables"""
        orig_data = super(CloudVPCTemplate, self).manager_userdata_head()
        data = [
            "export DOCKERCLOUD_USER='", Ref("DockerCloudUsername"), "'\n",
            "export DOCKERCLOUD_API_KEY='", Ref("DockerCloudAPIKey"), "'\n",
            "export SWARM_NAME='", Ref("DockerCloudClusterName"), "'\n",
            "export INTERNAL_ENDPOINT='", GetAtt("ExternalLoadBalancer", "DNSName"), "'\n",
            "export DOCKERCLOUD_REST_HOST='", Ref("DockerCloudRestHost"), "'\n",
            "export DOCKERCLOUD_ID_HOST='", Ref("DockerCloudIDHost"), "'\n",
        ]
        return orig_data + data

    def manager_userdata_body(self):
        """ This is the body of the userdata """
        orig_data = super(CloudVPCTemplate, self).manager_userdata_body()
        script_dir = path.dirname(__file__)

        cloud_path = path.relpath("data/cloud/manager_node_userdata.sh")
        cloud_file_path = path.join(script_dir, cloud_path)
        cloud_data = resources.userdata_from_file(cloud_file_path)

        return orig_data + cloud_data

    def add_outputs(self):
        """
        "ConnectToThisCluster" : {
             "Condition" : "DockerCloudRegistration",
             "Description" : "Use this command to manage this swarm cluster from your local Docker Engine.",
             "Value" : {
                 "Fn::Join": [ "", ["docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_HOST dockercloud/client ", { "Ref": "DockerCloudClusterName" } ] ]
             }
        },
        """
        super(CloudVPCTemplate, self).add_outputs()
        self.template.add_output(Output(
            "ConnectToThisCluster",
            Description="Use this command to manage this swarm cluster from your local Docker Engine.",
            Condition="DockerCloudRegistration",
            Value=Join(
                "",
                ["docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_HOST dockercloud/client ",
                 Ref("DockerCloudClusterName")])
        ))

    def add_conditions(self):
        """
        "DockerCloudRegistration": {
             "Fn::And": [
                 {"Fn::Not": [{"Fn::Equals" : [{"Ref" : "DockerCloudClusterName"}, ""]}]},
                 {"Fn::Not": [{"Fn::Equals" : [{"Ref" : "DockerCloudUsername"}, ""]}]},
                 {"Fn::Not": [{"Fn::Equals" : [{"Ref" : "DockerCloudAPIKey"}, ""]}]}
             ]
        }
        """
        super(CloudVPCTemplate, self).add_conditions()
        self.template.add_condition(
            "DockerCloudRegistration",
            And(
                Not(Equals(Ref("DockerCloudClusterName"), "")),
                Not(Equals(Ref("DockerCloudUsername"), "")),
                Not(Equals(Ref("DockerCloudAPIKey"), "")),
            )
        )

    def parameter_groups(self):
        parameter_groups = super(CloudVPCTemplate, self).parameter_groups()
        parameter_groups.append(
            {"Label": {"default": "Docker Cloud registration (optional)"},
             "Parameters": ["DockerCloudClusterName", "DockerCloudUsername",
                            "DockerCloudAPIKey", "DockerCloudRestHost"]}
        )
        return parameter_groups

    def add_cloud_clustername(self):
        self.template.add_parameter(Parameter(
            "DockerCloudClusterName",
            Description="Name of the cluster (namespace/cluster_name) to "
                        "be used when registering this Swarm with Docker Cloud",
            Type='String',
            ConstraintDescription="Must be in the format 'namespace/cluster_name' and "
                                  "must only contain letters, digits and hyphens",
            AllowedPattern="([a-z0-9]+/[a-z0-9-]+)?"
        ))
        self.add_to_parameters(('DockerCloudClusterName', {"default": "Swarm name?"}))

    def add_cloud_username(self):
        self.template.add_parameter(Parameter(
            "DockerCloudUsername",
            Description="Docker ID username to use during registration",
            Type='String',
            ConstraintDescription="Must only contain letters or digits",
            AllowedPattern="([a-z0-9]+)?"
        ))
        self.add_to_parameters(('DockerCloudUsername', {"default": "Docker ID Username?"}))

    def add_cloud_apikey(self):
        self.template.add_parameter(Parameter(
            "DockerCloudAPIKey",
            Description="Docker ID API key to use during registration",
            Type='String',
            NoEcho=True
        ))
        self.add_to_parameters(('DockerCloudAPIKey', {"default": "Docker ID API key?"}))

    def add_cloud_rest_host(self):
        self.template.add_parameter(Parameter(
            "DockerCloudRestHost",
            Description="Docker Cloud environment",
            Type='String',
            Default="https://cloud.docker.com"
        ))
        self.add_to_parameters(('DockerCloudRestHost', {"default": "Docker Cloud environment?"}))

    def add_cloud_id_host(self):
        self.template.add_parameter(Parameter(
            "DockerCloudIDHost",
            Description="ID service environment",
            Type='String',
            Default="https://id.docker.com"
        ))
        self.add_to_parameters(('DockerCloudIDHost', {"default": "ID service environment?"}))


class CloudVPCExistingTemplate(CloudVPCTemplate, ExistingVPCTemplate):
    """ Cloud Template for existing VPC."""
    def __init__(self, docker_version, edition_version,
                 docker_for_aws_version, channel, amis,
                 create_vpc=False, template_description=None):
        super(CloudVPCExistingTemplate, self).__init__(
            docker_version, edition_version,
            docker_for_aws_version,
            channel, amis,
            create_vpc=create_vpc,
            template_description=template_description
            )

from troposphere import FindInMap, Ref, Join, Select, GetAZs, Tags
from troposphere.efs import FileSystem, MountTarget


def add_resource_efs(template):
    for perfmode in ["GP", "MaxIO"]:
        tags = Tags(Name=Join("-", [Ref("AWS::StackName"), "EFS-" + perfmode]))
        template.add_resource(
            FileSystem('FileSystem' + perfmode,
                Condition="EFSSupported",
                PerformanceMode=perfmode,
                FileSystemTags=tags
            )
        )

def add_resource_mount_targets(template):
    for perfmode in ["GP", "MaxIO"]:
        for az in ["1", "2", "3"]:
            template.add_resource(
                MountTarget('MountTarget' + perfmode + az,
                    FileSystemId=Ref("FileSystem" + perfmode),
                    SecurityGroups=Ref("SwarmWideSG"),
                    SubnetId=Ref("PubSubnetAz" + az)
                )
            )

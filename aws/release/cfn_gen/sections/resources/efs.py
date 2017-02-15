from troposphere import FindInMap, Ref, Join, Select, GetAZs, Tags
from troposphere.efs import FileSystem, MountTarget


def add_resource_efs(template):
    perfmode_strings = {'GP': 'generalPurpose', 'MaxIO': 'maxIO'}
    for perfmode in ["GP", "MaxIO"]:
        tags = Tags(Name=Join("-", [Ref("AWS::StackName"), "EFS-" + perfmode]))
        template.add_resource(
            FileSystem('FileSystem' + perfmode,
                Condition="EFSSupported",
                PerformanceMode=perfmode_strings[perfmode],
                FileSystemTags=tags
            )
        )

def add_resource_mount_targets(template):
    for perfmode in ["GP", "MaxIO"]:
        for az in ["1", "2", "3"]:
            template.add_resource(
                MountTarget('MountTarget' + perfmode + az,
                    DependsOn=["FileSystem" + perfmode, "SwarmWideSG"],
                    Condition="EFSSupported",
                    FileSystemId=Ref("FileSystem" + perfmode),
                    SecurityGroups=[Ref("SwarmWideSG")],
                    SubnetId=Ref("PubSubnetAz" + az)
                )
            )

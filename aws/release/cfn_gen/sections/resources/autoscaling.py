from troposphere import FindInMap, GetAtt, Ref, If, Join
from troposphere.autoscaling import LifecycleHook, AutoScalingGroup, LaunchConfiguration
from troposphere.policies import (
    AutoScalingRollingUpdate, UpdatePolicy, CreationPolicy, ResourceSignal)
from troposphere.ec2 import BlockDeviceMapping, EBSBlockDevice


def add_resource_manager_upgrade_hook(template):
    """
    "SwarmManagerUpgradeHook": {
        "Type": "AWS::AutoScaling::LifecycleHook",
        "DependsOn" : "SwarmSQS",
        "Properties": {
            "AutoScalingGroupName": { "Ref": "ManagerAsg" },
            "LifecycleTransition": "autoscaling:EC2_INSTANCE_TERMINATING",
            "NotificationTargetARN": { "Fn::GetAtt": [ "SwarmSQS", "Arn" ] },
            "RoleARN": { "Fn::GetAtt": [ "ProxyRole", "Arn" ] }
        }
    },
    """
    template.add_resource(LifecycleHook(
        "SwarmManagerUpgradeHook",
        DependsOn="SwarmSQS",
        AutoScalingGroupName=Ref("ManagerAsg"),
        LifecycleTransition="autoscaling:EC2_INSTANCE_TERMINATING",
        NotificationTargetARN=GetAtt("SwarmSQS", "Arn"),
        RoleARN=GetAtt("ProxyRole", "Arn")
    ))


def add_resource_worker_upgrade_hook(template):
    """
    "SwarmWorkerUpgradeHook": {
        "Type": "AWS::AutoScaling::LifecycleHook",
        "DependsOn" : "SwarmSQS",
        "Properties": {
            "AutoScalingGroupName": { "Ref": "NodeAsg" },
            "LifecycleTransition": "autoscaling:EC2_INSTANCE_TERMINATING",
            "NotificationTargetARN": { "Fn::GetAtt": [ "SwarmSQS", "Arn" ] },
            "RoleARN": { "Fn::GetAtt": [ "WorkerRole", "Arn" ] }
        }
    },
    """
    template.add_resource(LifecycleHook(
        "SwarmWorkerUpgradeHook",
        DependsOn="SwarmSQS",
        AutoScalingGroupName=Ref("NodeAsg"),
        LifecycleTransition="autoscaling:EC2_INSTANCE_TERMINATING",
        NotificationTargetARN=GetAtt("SwarmSQS", "Arn"),
        RoleARN=GetAtt("WorkerRole", "Arn")
    ))


def add_resource_manager_autoscalegroup(template, create_vpc,
                                        launch_config_name, lb_list,
                                        health_check_grace_period=300):
    """
    "ManagerAsg" : {
        "DependsOn" : ["SwarmDynDBTable", "PubSubnetAz1",
            "PubSubnetAz2", "PubSubnetAz3", "ExternalLoadBalancer"],
        "Type" : "AWS::AutoScaling::AutoScalingGroup",
        "Properties" : {
            "VPCZoneIdentifier" : [{
                "Fn::If": [
                  "HasOnly2AZs",
                    { "Fn::Join" : [",", [ { "Ref" : "PubSubnetAz1" },
                        { "Ref" : "PubSubnetAz2" } ] ] },
                    { "Fn::Join" : [",", [ { "Ref" : "PubSubnetAz1" },
                        { "Ref" : "PubSubnetAz2" }, { "Ref" : "PubSubnetAz3" } ] ] }
                ]
            }],
            "LaunchConfigurationName" : { "Ref" : "ManagerLaunchConfigBeta3" },
            "LoadBalancerNames" : [ { "Ref" : "ExternalLoadBalancer" } ],
            "MinSize" : "0",
            "MaxSize" : "5",
            "HealthCheckType": "ELB",
            "HealthCheckGracePeriod": "300",
            "DesiredCapacity" : { "Ref" : "ManagerSize" },
            "Tags": [
                { "Key" : "Name",
                  "Value" : { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName"}, "Manager" ] ] },
                  "PropagateAtLaunch" : "true" },
                { "Key" : "swarm-node-type",
                  "Value" : "manager",
                  "PropagateAtLaunch" : "true" },
                { "Key" : "swarm-stack-id",
                  "Value" : { "Ref" : "AWS::StackId"},
                  "PropagateAtLaunch" : "true" },
                { "Key": "DOCKER_FOR_AWS_VERSION",
                  "Value": { "Fn::FindInMap" : [ "DockerForAWS", "version", "forAws" ] },
                  "PropagateAtLaunch" : "true" },
                { "Key": "DOCKER_VERSION",
                 "Value": { "Fn::FindInMap" : [ "DockerForAWS", "version", "docker" ] },
                 "PropagateAtLaunch" : "true"
                }
            ]
        },
        "CreationPolicy": {
            "ResourceSignal": {
              "Count": { "Ref" : "ManagerSize"},
              "Timeout": "PT20M"
            }
        },
        "UpdatePolicy" : {
          "AutoScalingRollingUpdate" : {
             "MinInstancesInService" : { "Ref" : "ManagerSize"},
             "MaxBatchSize" : "1",
             "WaitOnResourceSignals" : "true",
             "PauseTime" : "PT20M"
          }
       }
    },
    """
    elb_ref_list = []
    for lb in lb_list:
        elb_ref_list.append(Ref(lb))

    if create_vpc:
        depends = ["SwarmDynDBTable", "PubSubnetAz1",
                   "PubSubnetAz2", "PubSubnetAz3"]
    else:
        depends = ["SwarmDynDBTable", ]

    # add the ELBs as deps
    depends.extend(lb_list)

    template.add_resource(AutoScalingGroup(
        "ManagerAsg",
        DependsOn=depends,
        Tags=[
            {'Key': "Name",
             'Value': Join("-", [Ref("AWS::StackName"), "Manager"]),
             'PropagateAtLaunch': True},
            {'Key': "swarm-node-type",
             'Value': "manager",
             'PropagateAtLaunch': True},
            {'Key': "swarm-stack-id",
             'Value': Ref("AWS::StackId"),
             'PropagateAtLaunch': True},
            {'Key': "DOCKER_FOR_AWS_VERSION",
             'Value': FindInMap("DockerForAWS", "version", "forAws"),
             'PropagateAtLaunch': True},
            {'Key': "DOCKER_VERSION",
             'Value': FindInMap("DockerForAWS", "version", "docker"),
             'PropagateAtLaunch': True}
        ],
        LaunchConfigurationName=Ref(launch_config_name),
        MinSize=0,
        MaxSize=5,
        DesiredCapacity=Ref("ManagerSize"),
        VPCZoneIdentifier=[
            If("HasOnly2AZs",
               Join(",", [Ref("PubSubnetAz1"), Ref("PubSubnetAz2")]),
               Join(",", [Ref("PubSubnetAz1"), Ref("PubSubnetAz2"), Ref("PubSubnetAz3")]))
        ],
        LoadBalancerNames=elb_ref_list,
        HealthCheckType="ELB",
        HealthCheckGracePeriod=health_check_grace_period,
        UpdatePolicy=UpdatePolicy(
            AutoScalingRollingUpdate=AutoScalingRollingUpdate(
                PauseTime='PT20M',
                MinInstancesInService=Ref("ManagerSize"),
                MaxBatchSize='1',
                WaitOnResourceSignals=True
            )
        ),
        CreationPolicy=CreationPolicy(
            ResourceSignal=ResourceSignal(
                Timeout='PT20M',
                Count=Ref("ManagerSize")
                )
        )
    ))


def add_resource_worker_autoscalegroup(template, launch_config_name):
    """
    "NodeAsg" : {
        "DependsOn" : "ManagerAsg",
        "Type" : "AWS::AutoScaling::AutoScalingGroup",
        "Properties" : {
            "VPCZoneIdentifier" : [{
                "Fn::If": [
                  "HasOnly2AZs",
                    { "Fn::Join" : [",", [ { "Ref" : "PubSubnetAz1" },
                        { "Ref" : "PubSubnetAz2" } ] ] },
                    { "Fn::Join" : [",", [ { "Ref" : "PubSubnetAz1" },
                        { "Ref" : "PubSubnetAz2" }, { "Ref" : "PubSubnetAz3" } ] ] }
                ]
            }],
            "LaunchConfigurationName" : { "Ref" : "NodeLaunchConfigBeta3" },
            "LoadBalancerNames" : [ { "Ref" : "ExternalLoadBalancer" } ],
            "MinSize" : "0",
            "MaxSize" : "1000",
            "HealthCheckType": "ELB",
            "HealthCheckGracePeriod": "300",
            "DesiredCapacity" : { "Ref" : "ClusterSize"},
            "Tags": [
                { "Key" : "Name",
                  "Value" : { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName"}, "worker" ] ] },
                  "PropagateAtLaunch" : "true" },
                { "Key" : "swarm-node-type",
                  "Value" : "worker",
                  "PropagateAtLaunch" : "true" },
                { "Key" : "swarm-stack-id",
                  "Value" : { "Ref" : "AWS::StackId"},
                  "PropagateAtLaunch" : "true" },
                { "Key": "DOCKER_FOR_AWS_VERSION",
                  "Value": { "Fn::FindInMap" : [ "DockerForAWS", "version", "forAws" ] },
                  "PropagateAtLaunch" : "true" },
                { "Key": "DOCKER_VERSION",
                  "Value": { "Fn::FindInMap" : [ "DockerForAWS", "version", "docker" ] },
                  "PropagateAtLaunch" : "true"
                }
            ]
        },
        "CreationPolicy": {
            "ResourceSignal": {
              "Count": { "Ref" : "ClusterSize"},
              "Timeout": "PT2H"
            }
        },
        "UpdatePolicy" : {
          "AutoScalingRollingUpdate" : {
             "MinInstancesInService" : { "Ref" : "ClusterSize"},
             "MaxBatchSize" : "1",
             "WaitOnResourceSignals" : "true",
             "PauseTime" : "PT1H"
          }
       }
    },
    """
    template.add_resource(AutoScalingGroup(
        "NodeAsg",
        DependsOn="ManagerAsg",
        Tags=[
            {'Key': "Name",
             'Value': Join("-", [Ref("AWS::StackName"), "worker"]),
             'PropagateAtLaunch': True},
            {'Key': "swarm-node-type",
             'Value': "worker",
             'PropagateAtLaunch': True},
            {'Key': "swarm-stack-id",
             'Value': Ref("AWS::StackId"),
             'PropagateAtLaunch': True},
            {'Key': "DOCKER_FOR_AWS_VERSION",
             'Value': FindInMap("DockerForAWS", "version", "forAws"),
             'PropagateAtLaunch': True},
            {'Key': "DOCKER_VERSION",
             'Value': FindInMap("DockerForAWS", "version", "docker"),
             'PropagateAtLaunch': True}
        ],
        LaunchConfigurationName=Ref(launch_config_name),
        MinSize=0,
        MaxSize=1000,
        DesiredCapacity=Ref("ClusterSize"),
        VPCZoneIdentifier=[
            If("HasOnly2AZs",
               Join(",", [Ref("PubSubnetAz1"), Ref("PubSubnetAz2")]),
               Join(",", [Ref("PubSubnetAz1"), Ref("PubSubnetAz2"), Ref("PubSubnetAz3")]))
        ],
        LoadBalancerNames=[Ref("ExternalLoadBalancer")],
        HealthCheckType="ELB",
        HealthCheckGracePeriod=300,
        UpdatePolicy=UpdatePolicy(
            AutoScalingRollingUpdate=AutoScalingRollingUpdate(
                PauseTime='PT1H',
                MinInstancesInService=Ref("ClusterSize"),
                MaxBatchSize='1',
                WaitOnResourceSignals=True
            )
        ),
        CreationPolicy=CreationPolicy(
            ResourceSignal=ResourceSignal(
                Timeout='PT2H',
                Count=Ref("ClusterSize")
                )
        )
    ))


def add_resource_manager_launch_config(template, user_data, launch_config_name="ManagerLaunchConfig"):
    """
    "ManagerLaunchConfigBeta3": {
        "DependsOn": "ExternalLoadBalancer",
        "Type": "AWS::AutoScaling::LaunchConfiguration",
        "Properties": {
            "InstanceType": {"Ref" : "ManagerInstanceType"},
            "BlockDeviceMappings" : [ {
                "DeviceName" : "/dev/xvdb",
                "Ebs" : {
                    "VolumeSize" : { "Ref" : "ManagerDiskSize" },
                    "VolumeType" : { "Ref" : "ManagerDiskType" }
                }
             }],
            "IamInstanceProfile" : { "Ref" : "ProxyInstanceProfile" },
            "KeyName": {
                "Ref": "KeyName"
            },
            "ImageId": {
                "Fn::FindInMap": ["AWSRegionArch2AMI", {
                    "Ref": "AWS::Region"
                }, {
                    "Fn::FindInMap": ["AWSInstanceType2Arch", {"Ref" : "ManagerInstanceType"}, "Arch"]
                }]
            },
            "AssociatePublicIpAddress": "true",
            "SecurityGroups": [ { "Ref" : "ManagerVpcSG"}, { "Ref" : "SwarmWideSG"} ],
            "UserData": {
                "Fn::Base64": {
                    "Fn::Join": [
                        "", [
                            "#!/bin/sh\n",
                            "echo \"", {"Ref": "ExternalLoadBalancer"}, "\" > /var/lib/docker/editions/lb_name\n",
                            "echo \"# hostname : ELB_name\" >> /var/lib/docker/editions/elb.config\n",
                            "echo \"127.0.0.1: ", {"Ref": "ExternalLoadBalancer"}, "\" >> /var/lib/docker/editions/elb.config\n",
                            "echo \"localhost: ", {"Ref": "ExternalLoadBalancer"}, "\" >> /var/lib/docker/editions/elb.config\n",
                            "echo \"default: ", {"Ref": "ExternalLoadBalancer"}, "\" >> /var/lib/docker/editions/elb.config\n",
                            "export DOCKER_FOR_IAAS_VERSION='", { "Fn::FindInMap" : [ "DockerForAWS", "version", "forAws" ] }, "'\n",
                            "export LOCAL_IP=$(wget -qO- http://169.254.169.254/latest/meta-data/local-ipv4)\n",
                            "export ENABLE_CLOUDWATCH_LOGS='", {"Ref": "EnableCloudWatchLogs"} , "'\n",
                            "if [ $ENABLE_CLOUDWATCH_LOGS == 'yes' ] ; then \n",
                                "echo '{\"experimental\": true, \"log-driver\": \"awslogs\",\"log-opts\": {\"awslogs-group\":\"",
                                { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName"}, "lg" ] ] },
                                "\", \"tag\": \"{{.Name}}-{{.ID}}\" }}' > /etc/docker/daemon.json \n",
                            "else\n",
                                "echo '{\"experimental\": true }' > /etc/docker/daemon.json \n",
                            "fi\n",
                            "chown -R docker /home/docker/\n",
                            "chgrp -R docker /home/docker/\n",
                            "rc-service docker restart\n",
                            "sleep 5\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file --name=meta-aws --restart=always -d -p $LOCAL_IP:9024:8080 ",
                            "-e AWS_REGION='",{ "Ref" : "AWS::Region" }, "' ",
                            "-e MANAGER_SECURITY_GROUP_ID='",{ "Ref" : "ManagerVpcSG" }, "' ",
                            "-e WORKER_SECURITY_GROUP_ID='",{ "Ref" : "NodeVpcSG" }, "' ",
                            "-v /var/run/docker.sock:/var/run/docker.sock ",
                            "docker4x/meta-aws:$DOCKER_FOR_IAAS_VERSION metaserver -iaas_provider=aws\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file --restart=no -d ",
                            "-e DYNAMODB_TABLE='", { "Ref" : "SwarmDynDBTable" } , "' ",
                            "-e NODE_TYPE='manager' ",
                            "-e REGION='",{ "Ref" : "AWS::Region" }, "' ",
                            "-e STACK_NAME='",{ "Ref" : "AWS::StackName" }, "' ",
                            "-e STACK_ID='",{ "Ref" : "AWS::StackId" }, "' ",
                            "-e ACCOUNT_ID='",{ "Ref" : "AWS::AccountId" }, "' ",
                            "-e INSTANCE_NAME='ManagerAsg' ",
                            "-e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION ",
                            "-v /var/run/docker.sock:/var/run/docker.sock ",
                            "-v /usr/bin/docker:/usr/bin/docker ",
                            "-v /var/log:/var/log ",
                            "docker4x/init-aws:$DOCKER_FOR_IAAS_VERSION\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file --name=guide-aws --restart=always -d ",
                            "-e DYNAMODB_TABLE='", { "Ref" : "SwarmDynDBTable" } , "' ",
                            "-e NODE_TYPE='manager' ",
                            "-e REGION='",{ "Ref" : "AWS::Region" }, "' ",
                            "-e STACK_NAME='",{ "Ref" : "AWS::StackName" }, "' ",
                            "-e INSTANCE_NAME='ManagerAsg' ",
                            "-e VPC_ID='",{ "Ref" : "Vpc" }, "' ",
                            "-e STACK_ID='",{ "Ref" : "AWS::StackId" }, "' ",
                            "-e ACCOUNT_ID='",{ "Ref" : "AWS::AccountId" }, "' ",
                            "-e SWARM_QUEUE='",{ "Ref" : "SwarmSQS" }, "' ",
                            "-e CLEANUP_QUEUE='",{ "Ref" : "SwarmSQSCleanup" }, "' ",
                            "-e RUN_VACUUM='",{ "Ref" : "EnableSystemPrune" }, "' ",
                            "-e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION ",
                            "-v /var/run/docker.sock:/var/run/docker.sock ",
                            "-v /usr/bin/docker:/usr/bin/docker ",
                            "docker4x/guide-aws:$DOCKER_FOR_IAAS_VERSION\n",

                            "docker volume create --name sshkey\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file -ti --rm ",
                            "--user root ",
                            "-v sshkey:/etc/ssh ",
                            "--entrypoint ssh-keygen ",
                            "docker4x/shell-aws:$DOCKER_FOR_IAAS_VERSION ",
                            "-A\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file --name=shell-aws --restart=always -d -p 22:22 ",
                            "-v /home/docker/:/home/docker/ ",
                            "-v /var/run/docker.sock:/var/run/docker.sock ",
                            "-v /usr/bin/docker:/usr/bin/docker ",
                            "-v /var/log:/var/log ",
                            "-v sshkey:/etc/ssh ",
                            "-v /etc/passwd:/etc/passwd:ro ",
                            "-v /etc/shadow:/etc/shadow:ro ",
                            "-v /etc/group:/etc/group:ro ",
                            "docker4x/shell-aws:$DOCKER_FOR_IAAS_VERSION\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file --name=l4controller-aws --restart=always -d ",
                            "-v /var/run/docker.sock:/var/run/docker.sock ",
                            "-v /var/lib/docker/editions:/var/lib/docker/editions ",
                            "docker4x/l4controller-aws:$DOCKER_FOR_IAAS_VERSION run --log=4 --all=true\n"
                        ]
                    ]
                }
            }
        }
    },
    """
    template.add_resource(LaunchConfiguration(
        launch_config_name,
        DependsOn="ExternalLoadBalancer",
        UserData=user_data,  # compile_manager_node_userdata(),
        ImageId=FindInMap(
            "AWSRegionArch2AMI",
            Ref("AWS::Region"),
            FindInMap("AWSInstanceType2Arch", Ref("ManagerInstanceType"), "Arch")
        ),
        KeyName=Ref("KeyName"),
        BlockDeviceMappings=[
            BlockDeviceMapping(
                DeviceName="/dev/xvdb",
                Ebs=EBSBlockDevice(
                    VolumeSize=Ref("ManagerDiskSize"),
                    VolumeType=Ref("ManagerDiskType")
                )
            ),
        ],
        SecurityGroups=[Ref("ManagerVpcSG"), Ref("SwarmWideSG")],
        InstanceType=Ref("ManagerInstanceType"),
        AssociatePublicIpAddress=True,
        IamInstanceProfile=Ref("ProxyInstanceProfile"),
    ))


def add_resource_worker_launch_config(template, user_data, launch_config_name="NodeLaunchConfig"):
    """
    "NodeLaunchConfigBeta3": {
        "DependsOn": "ManagerAsg",
        "Type": "AWS::AutoScaling::LaunchConfiguration",
        "Properties": {
            "InstanceType": {"Ref" : "InstanceType"},
            "BlockDeviceMappings" : [ {
                "DeviceName" : "/dev/xvdb",
                "Ebs" : {
                    "VolumeSize" : { "Ref" : "WorkerDiskSize" },
                    "VolumeType" : { "Ref" : "WorkerDiskType" }
                }
             }],
            "IamInstanceProfile" : { "Ref" : "WorkerInstanceProfile" },
            "KeyName": {
                "Ref": "KeyName"
            },
            "ImageId": {
                "Fn::FindInMap": ["AWSRegionArch2AMI", {
                    "Ref": "AWS::Region"
                }, {
                    "Fn::FindInMap": ["AWSInstanceType2Arch", {"Ref" : "InstanceType"}, "Arch"]
                }]
            },
            "AssociatePublicIpAddress": "true",
            "SecurityGroups": [ { "Ref" : "NodeVpcSG"} ],
            "UserData": {
                "Fn::Base64": {
                    "Fn::Join": [
                        "", [
                            "#!/bin/sh\n",
                            "export DOCKER_FOR_IAAS_VERSION='", { "Fn::FindInMap" : [ "DockerForAWS", "version", "forAws" ] }, "'\n",
                            "export ENABLE_CLOUDWATCH_LOGS='", {"Ref": "EnableCloudWatchLogs"} , "'\n",
                            "if [ $ENABLE_CLOUDWATCH_LOGS == 'yes' ] ; then \n",
                                "echo '{\"experimental\": true, \"log-driver\": \"awslogs\",\"log-opts\": {\"awslogs-group\":\"",
                                { "Fn::Join": [ "-", [ { "Ref": "AWS::StackName"}, "lg" ] ] },
                                "\", \"tag\": \"{{.Name}}-{{.ID}}\" }}' > /etc/docker/daemon.json \n",
                            "else\n",
                                "echo '{\"experimental\": true }' > /etc/docker/daemon.json \n",
                            "fi\n",
                            "chown -R docker /home/docker/\n",
                            "chgrp -R docker /home/docker/\n",
                            "rc-service docker restart\n",
                            "sleep 5\n",
                            "docker run --label com.docker.editions.system --log-driver=json-file --restart=no -d ",
                            "-e DYNAMODB_TABLE='", { "Ref" : "SwarmDynDBTable" } , "' ",
                            "-e NODE_TYPE='worker' ",
                            "-e REGION='",{ "Ref" : "AWS::Region" }, "' ",
                            "-e STACK_NAME='",{ "Ref" : "AWS::StackName" }, "' ",
                            "-e STACK_ID='",{ "Ref" : "AWS::StackId" }, "' ",
                            "-e ACCOUNT_ID='",{ "Ref" : "AWS::AccountId" }, "' ",
                            "-e INSTANCE_NAME='NodeAsg' ",
                            "-e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION ",
                            "-v /var/run/docker.sock:/var/run/docker.sock ",
                            "-v /usr/bin/docker:/usr/bin/docker ",
                            "-v /var/log:/var/log ",
                            "docker4x/init-aws:$DOCKER_FOR_IAAS_VERSION\n",

                            "docker volume create --name sshkey\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file -ti --rm ",
                            "--user root ",
                            "-v sshkey:/etc/ssh ",
                            "--entrypoint ssh-keygen ",
                            "docker4x/shell-aws:$DOCKER_FOR_IAAS_VERSION ",
                            "-A\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file --name=shell-aws --restart=always -d -p 22:22 ",
                            "-v /home/docker/:/home/docker/ ",
                            "-v /var/run/docker.sock:/var/run/docker.sock ",
                            "-v /usr/bin/docker:/usr/bin/docker ",
                            "-v /var/log:/var/log ",
                            "-v sshkey:/etc/ssh ",
                            "-v /etc/passwd:/etc/passwd:ro ",
                            "-v /etc/shadow:/etc/shadow:ro ",
                            "-v /etc/group:/etc/group:ro ",
                            "docker4x/shell-aws:$DOCKER_FOR_IAAS_VERSION\n",

                            "docker run --label com.docker.editions.system --log-driver=json-file --name=guide-aws --restart=always -d ",
                            "-e DYNAMODB_TABLE='", { "Ref" : "SwarmDynDBTable" } , "' ",
                            "-e NODE_TYPE='worker' ",
                            "-e REGION='",{ "Ref" : "AWS::Region" }, "' ",
                            "-e STACK_NAME='",{ "Ref" : "AWS::StackName" }, "' ",
                            "-e INSTANCE_NAME='NodeAsg' ",
                            "-e VPC_ID='",{ "Ref" : "Vpc" }, "' ",
                            "-e STACK_ID='",{ "Ref" : "AWS::StackId" }, "' ",
                            "-e ACCOUNT_ID='",{ "Ref" : "AWS::AccountId" }, "' ",
                            "-e SWARM_QUEUE='",{ "Ref" : "SwarmSQS" }, "' ",
                            "-e CLEANUP_QUEUE='",{ "Ref" : "SwarmSQSCleanup" }, "' ",
                            "-e RUN_VACUUM='",{ "Ref" : "EnableSystemPrune" }, "' ",
                            "-e DOCKER_FOR_IAAS_VERSION=$DOCKER_FOR_IAAS_VERSION ",
                            "-v /var/run/docker.sock:/var/run/docker.sock ",
                            "-v /usr/bin/docker:/usr/bin/docker ",
                            "docker4x/guide-aws:$DOCKER_FOR_IAAS_VERSION\n"
                        ]
                    ]
                }
            }
        }
    },
    """
    template.add_resource(LaunchConfiguration(
        launch_config_name,  # TODO: dynamic
        DependsOn="ManagerAsg",
        UserData=user_data,  # compile_worker_node_userdata(),
        ImageId=FindInMap(
            "AWSRegionArch2AMI",
            Ref("AWS::Region"),
            FindInMap("AWSInstanceType2Arch", Ref("InstanceType"), "Arch")
        ),
        KeyName=Ref("KeyName"),
        BlockDeviceMappings=[
            BlockDeviceMapping(
                DeviceName="/dev/xvdb",
                Ebs=EBSBlockDevice(
                    VolumeSize=Ref("WorkerDiskSize"),
                    VolumeType=Ref("WorkerDiskType")
                )
            ),
        ],
        SecurityGroups=[Ref("NodeVpcSG")],
        InstanceType=Ref("InstanceType"),
        AssociatePublicIpAddress=True,
        IamInstanceProfile=Ref("WorkerInstanceProfile"),
    ))

from os import path
from troposphere import Join, Ref, FindInMap


def common_userdata_head():
    """ The Head of the userdata script, this is where
    you would declare all of your shell variables"""
    data = [
        "export EXTERNAL_LB='", Ref("ExternalLoadBalancer"), "'\n",
        "export DOCKER_FOR_IAAS_VERSION='", FindInMap("DockerForAWS", "version", "forAws"), "'\n",
        "export LOCAL_IP=$(wget -qO- http://169.254.169.254/latest/meta-data/local-ipv4)\n",
        "export INSTANCE_TYPE=$(wget -qO- http://169.254.169.254/latest/meta-data/instance-type)\n",
        "export NODE_AZ=$(wget -qO- http://169.254.169.254/latest/meta-data/placement/availability-zone/)\n",
        "export NODE_REGION=$(echo $NODE_AZ | sed 's/.$//')\n",
        "export ENABLE_CLOUDWATCH_LOGS='", Ref("EnableCloudWatchLogs"), "'\n",
        "export AWS_REGION='", Ref("AWS::Region"), "'\n",
        "export MANAGER_SECURITY_GROUP_ID='", Ref("ManagerVpcSG"), "'\n",
        "export WORKER_SECURITY_GROUP_ID='", Ref("NodeVpcSG"), "'\n",
        "export DYNAMODB_TABLE='", Ref("SwarmDynDBTable"), "'\n",
        "export STACK_NAME='", Ref("AWS::StackName"), "'\n",
        "export STACK_ID='", Ref("AWS::StackId"), "'\n",
        "export ACCOUNT_ID='", Ref("AWS::AccountId"), "'\n",
        "export VPC_ID='", Ref("Vpc"), "'\n",
        "export SWARM_QUEUE='", Ref("SwarmSQS"), "'\n",
        "export CLEANUP_QUEUE='", Ref("SwarmSQSCleanup"), "'\n",
        "export RUN_VACUUM='", Ref("EnableSystemPrune"), "'\n",
        "export LOG_GROUP_NAME='", Join("-", [Ref("AWS::StackName"), "lg"]), "'\n",
    ]
    return data


def manager_node_userdata_head():
    """ The Head of the userdata script, this is where
    you would declare all of your shell variables"""
    data = [
        "export NODE_TYPE='manager'\n",
        "export INSTANCE_NAME='ManagerAsg'\n"
    ]
    return common_userdata_head() + data


def manager_node_userdata_body():
    """ This is the body of the userdata """
    script_dir = path.dirname(__file__)
    manager_rel_path = "../../data/base/manager_node_userdata.sh"
    manager_path = path.join(script_dir, manager_rel_path)
    manager_data = userdata_from_file(manager_path)

    common_path = "../../data/base/common_userdata.sh"
    common_file_path = path.join(script_dir, common_path)
    common_data = userdata_from_file(common_file_path)

    return common_data + manager_data


def worker_node_userdata_head():
    """ The Head of the userdata script, this is where
    you would declare all of your shell variables"""
    data = [
        "export NODE_TYPE='worker'\n",
        "export INSTANCE_NAME='NodeAsg'\n"
    ]
    return common_userdata_head() + data


def worker_node_userdata_body():
    """ This is the body of the userdata """
    script_dir = path.dirname(__file__)
    worker_rel_path = "../../data/base/worker_node_userdata.sh"
    worker_path = path.join(script_dir, worker_rel_path)
    worker_data = userdata_from_file(worker_path)

    common_path = "../../data/base/common_userdata.sh"
    common_file_path = path.join(script_dir, common_path)
    common_data = userdata_from_file(common_file_path)

    return common_data + worker_data


def userdata_from_file(filepath, blanklines=False):
    """
    Imports userdata from a file.
    :type filepath: string
    :param filepath
    The absolute path to the file.
    :type blanklines: boolean
    :param blanklines
    If blank lines shoud be ignored
    rtype: list
    :return a list of lines from the userdata
    """

    data = []

    try:
        with open(filepath, 'r') as f:
            for line in f:
                if blanklines and line.strip('\n\r ') == '':
                    continue

                data.append(line)
    except IOError:
        raise IOError(u'Error opening or reading file: {}'.format(filepath))

    return data

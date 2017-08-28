#!/usr/bin/env python
import os
import json
import time
from datetime import datetime
import boto
import urllib2
from boto import ec2
import re
from collections import (OrderedDict, namedtuple)

NOW = datetime.now()
NOW_STRING = NOW.strftime("%m_%d_%Y")

AWS_ACCESS_KEY_ID = os.environ.get('AWS_ACCESS_KEY_ID')
AWS_SECRET_ACCESS_KEY = os.environ.get('AWS_SECRET_ACCESS_KEY')
S3_BUCKET_NAME = os.getenv('S3_BUCKET', 'docker-ci-editions')
ARM_S3_BUCKET_NAME = os.getenv('UPLOAD_S3_BUCKET', 'docker-for-azure')
ARM_AWS_ACCESS_KEY_ID = os.getenv('UPLOAD_S3_KEY', AWS_ACCESS_KEY_ID)
ARM_AWS_SECRET_ACCESS_KEY = os.getenv('UPLOAD_S3_SECRET', AWS_SECRET_ACCESS_KEY)
EDITIONS_COMMIT = os.getenv('EDITIONS_COMMIT',"unknown-editions-commit")
JENKINS_BUILD = os.getenv('JENKINS_BUILD',"unknown-jenkins-build")
UCP_TAG = os.getenv('UCP_TAG',"2.2.0")
DTR_TAG = os.getenv('DTR_TAG',"2.3.0")



def str2bool(v):
  return v.lower() in ("yes", "true", "t", "1")

def is_json(arm_template):
  try:
    with open(arm_template) as data_file:
      json_object = json.load(data_file)
  except ValueError, e:
    return False
  return True

def buildCustomData(data_file):
  customData = []
  with open(data_file) as f:
    for line in f:
      m = re.match(r'(.*?)((?:parameters|variables)\([^\)]*\))(.*$)', line)
      if m:
        customData += ['\'' + m.group(1) + '\'',
              m.group(2),
              '\'' + m.group(3) + '\'',
              '\'\n\'']
      else:
        customData += ['\'' + line + '\'']
  return customData

def upload_rg_template(release_channel, arm_template_name, tempfile, arm_type=''):

    # upload to s3, make public, return s3 URL
    s3_host_name = u"https://{}.s3.amazonaws.com".format(ARM_S3_BUCKET_NAME)
    s3_base_path = u"azure/{}/{}".format(release_channel, EDITIONS_COMMIT)
    s3_path_build = u"{}/{}/{}.tmpl".format(s3_base_path, JENKINS_BUILD, arm_template_name)
    s3_path = u"{}/{}.tmpl".format(s3_base_path, arm_template_name)
    latest_name = "latest.tmpl"
    if arm_type:
        latest_name = "{}-latest.tmpl".format(arm_type)

    s3_path_latest = u"azure/{}/{}".format(release_channel, latest_name)
    s3_full_url = u"{}/{}".format(s3_host_name, s3_path_build)

    s3conn = boto.connect_s3(ARM_AWS_ACCESS_KEY_ID, ARM_AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(ARM_S3_BUCKET_NAME)

    # Upload template to commit + build number folder
    print(u"Upload ARM template to {} in {} s3 bucket".format(s3_path_build, ARM_S3_BUCKET_NAME))
    key = bucket.new_key(s3_path_build)
    key.set_metadata("Content-Type", "application/json")
    key.set_contents_from_filename(tempfile)
    key.set_acl("public-read")

    # Copy to the commit root folder as well
    print(u"Copy template from {} to {} s3 bucket".format(s3_path_build, s3_path))
    srckey = bucket.get_key(s3_path_build)
    dstkey = bucket.new_key(s3_path)
    srckey.copy(ARM_S3_BUCKET_NAME, dstkey, preserve_acl=True, validate_dst_bucket=True)

    if release_channel == 'nightly' or release_channel == 'ddc-nightly'  or release_channel == 'cloud-nightly':
        print("This is a nightly build, update the latest.tmpl file.")
        print(u"Upload ARM template to {} in {} s3 bucket".format(
            s3_path_latest, ARM_S3_BUCKET_NAME))
        key = bucket.new_key(s3_path_latest)
        key.set_metadata("Content-Type", "application/json")
        key.set_contents_from_filename(tempfile)
        key.set_acl("public-read")

    return s3_full_url

def publish_rg_template(release_channel, docker_for_azure_version):
    # upload to s3, make public, return s3 URL
    s3_host_name = u"https://{}.s3.amazonaws.com".format(ARM_S3_BUCKET_NAME)
    s3_path = u"azure/{}/{}".format(release_channel, docker_for_azure_version)

    print(u"Update the latest.tmpl file to the release of {} in {} channel.".format(docker_for_azure_version, release_channel))
    latest_name = "latest.tmpl"
    s3_path_latest = u"azure/{}/{}".format(release_channel, latest_name)
    s3_full_url = u"{}/{}".format(s3_host_name, s3_path_latest)

    s3conn = boto.connect_s3(ARM_AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
    bucket = s3conn.get_bucket(ARM_S3_BUCKET_NAME)

    print(u"Copy Azure template from {} to {} s3 bucket".format(s3_path, s3_path_latest))
    srckey = bucket.get_key(s3_path)
    dstkey = bucket.new_key(s3_path_latest)
    srckey.copy(ARM_S3_BUCKET_NAME, dstkey, preserve_acl=True, validate_dst_bucket=True)
    return s3_full_url

def create_rg_template(vhd_sku, vhd_version, offer_id, release_channel, docker_version,
                        docker_for_azure_version, edition_addon, arm_template,
                        arm_template_name, public_platform, portal_endpoint, 
                        resource_mgr_url, storage_suffix, storage_endpoint, 
                        active_dir_url, svc_mgmt_url, resource_mgr_vm_suffix):
    # check if file exists before opening.
    flat_edition_version = docker_for_azure_version.replace(" ", "").replace("_", "").replace("-", "")
    flat_edition_version_upper = flat_edition_version.capitalize()

    with open(arm_template) as data_file:
        data = json.load(data_file)

    data['variables']['Description'] = u"Docker for Azure {0}".format(docker_for_azure_version)
    data['variables']['docker'] = u"{}".format(docker_version)
    data['variables']['imageSku'] = vhd_sku
    data['variables']['imageVersion'] = vhd_version
    data['variables']['imageOffer'] = offer_id
    data['variables']['channel'] = release_channel
    data['variables']['storageAccountDNSSuffix'] = storage_suffix
    data['variables']['portalFQDN'] = portal_endpoint
    data['variables']['editionAddOn'] = edition_addon

    if public_platform:
        # for the Public Azure platform, all API endpoints/urls have defaults
        custom_data = buildCustomData('custom-data-public.sh')
    else:
        # for non Public platforms (Gov, China, Stacks) set the endpoints/urls 
        data['variables']['resMgrUrl'] = resource_mgr_url
        data['variables']['storageEndpoint'] = storage_endpoint
        data['variables']['adUrl'] = active_dir_url
        data['variables']['svcMgmtUrl'] = svc_mgmt_url
        custom_data = buildCustomData('custom-data.sh')

    data['variables']['customData'] = '[concat(' + ', '.join(custom_data) + ')]'
    variables = data.get('variables')
    if variables:
        new_variables = {
            "extlbname": "[concat(variables('lbPublicIpDnsName'), '.', variables('storageLocation'), '.', '" + resource_mgr_vm_suffix + "')]"
        }
        variables.update(new_variables)
    
    outdir = u"dist/azure/{}/{}".format(release_channel, docker_for_azure_version)
    # if the directory doesn't exist, create it.
    if not os.path.exists(outdir):
        os.makedirs(outdir)

    outfile = u"{}/{}.tmpl".format(outdir, arm_template_name)

    with open(outfile, 'w') as outf:
        json.dump(data, outf, indent=4, sort_keys=True)

    # TODO: json validate
    if is_json(outfile):
      print(u"ARM template created in {}".format(outfile))
    else:
      print(u"ERROR: ARM template invalid in {}".format(outfile))

    return outfile

# @TODO VERIFY CLOUD TEMPLATE
# @TODO IMPLEMENT DDC TEMPLATE
def create_rg_cloud_template(release_channel, docker_version,
                        docker_for_azure_version, edition_addon, arm_template,
                        storage_suffix, portal_endpoint, arm_template_name):
    with open(arm_template) as data_file:
        data = json.load(data_file)

    # Updated custom data for Managers and Workers
    custom_data = buildCustomData('custom-data-public.sh')
    custom_data_cloud = buildCustomData('custom-data-public_cloud.sh')
    data['variables']['customData'] = '[concat(' + ', '.join(custom_data) + ", '\n', " + ', '.join(custom_data_cloud) + ')]'
    data['variables']['channel'] = release_channel
    data['variables']['storageAccountDNSSuffix'] = storage_suffix
    data['variables']['portalFQDN'] = portal_endpoint
    data['variables']['editionAddOn'] = edition_addon

    parameters = data.get('parameters')
    if parameters:
        new_parameters = {
            "DockerCloudClusterName" : {
                "type": "string",
                "metadata": {
                    "description": "Name of the cluster (namespace/cluster_name) to be used when registering this Swarm with Docker Cloud\n\nMust be in the format 'namespace/cluster_name' and must only contain letters, digits and hyphens"
                }
            },
            "DockerCloudUsername" : {
                "type" : "string",
                "metadata": {
                    "description": "Docker ID username to use during registration\n\nMust only contain letters or digits"
                }
            },
            "DockerCloudAPIKey" : {
                "type" : "securestring",
                "metadata": {
                    "description": "Docker ID API key to use during registration"
                }
            },
            "DockerCloudRestHost" : {
                "type": "string",
                "defaultValue": "https://cloud.docker.com",
                "metadata": {
                    "description": "Docker Cloud rest API endpoint"
                }
            },
            "DockerIDJWTURL" : {
                "type": "string",
                "defaultValue": "https://id.docker.com/api/id/v1/authz/token",
                "metadata": {
                    "description": "ID JWT token service URL"
                }
            },
            "DockerIDJWKURL" : {
                "type": "string",
                "defaultValue": "https://id.docker.com/api/id/v1/authz/certs",
                "metadata": {
                    "description": "ID JWK certificate URL"
                }
            }
        }
        parameters.update(new_parameters)

    variables = data.get('variables')
    if variables:
        new_variables = {
            "dockerCloudClusterName" : "[parameters('DockerCloudClusterName')]",
            "dockerCloudUsername" : "[parameters('DockerCloudUsername')]",
            "dockerCloudAPIKey" : "[parameters('DockerCloudAPIKey')]",
            "dockerCloudRestHost": "[parameters('DockerCloudRestHost')]",
            "dockerIDJWTURL": "[parameters('DockerIDJWTURL')]",
            "dockerIDJWKURL": "[parameters('DockerIDJWKURL')]",
        }
        variables.update(new_variables)

    outputs = data.get('outputs')
    if outputs:
        new_outputs = {
            "ConnectToThisCluster": {
                "type": "string",
                "value": "[concat('docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -e DOCKER_HOST dockercloud/client ', variables('DockerCloudClusterName'))]"
            }
        }
        outputs.update(new_outputs)

    outdir = u"dist/azure/{}/{}".format(release_channel, docker_for_azure_version)
    # if the directory doesn't exist, create it.
    if not os.path.exists(outdir):
        os.makedirs(outdir)

    outfile = u"{}/{}.tmpl".format(outdir, arm_template_name)

    with open(outfile, 'w') as outf:
        json.dump(data, outf, indent=4, sort_keys=True)

    # TODO: json validate
    if is_json(outfile):
      print(u"ARM template created in {}".format(outfile))
    else:
      print(u"ERROR: ARM template invalid in {}".format(outfile))

    return outfile

def create_rg_ddc_template(vhd_sku, vhd_version, offer_id, release_channel, docker_version,
                        docker_for_azure_version, edition_addon, arm_template, 
                        arm_template_name, public_platform, portal_endpoint, 
                        resource_mgr_url, storage_suffix, storage_endpoint, 
                        active_dir_url, svc_mgmt_url, resource_mgr_vm_suffix):
    with open(arm_template) as data_file:
        data = json.load(data_file)

    # Update managers to 3,5,7

    # Updated custom data for Managers and Workers
    if public_platform:
        # for the Public Azure platform, all API endpoints/urls have defaults
        custom_data = buildCustomData('custom-data-public.sh')
        custom_data_ddc = buildCustomData('custom-data-public_ddc.sh')
    else:
        # for non Public platforms (Gov, China, Stacks) set the endpoints/urls 
        data['variables']['resMgrUrl'] = resource_mgr_url
        data['variables']['storageEndpoint'] = storage_endpoint
        data['variables']['adUrl'] = active_dir_url
        data['variables']['svcMgmtUrl'] = svc_mgmt_url
        custom_data = buildCustomData('custom-data.sh')
        custom_data_ddc = buildCustomData('custom-data_ddc.sh')

    data['variables']['customData'] = '[concat(' + ', '.join(custom_data) + ", '\n', " +  ', '.join(custom_data_ddc) + ')]'

    data['variables']['imageSku'] = vhd_sku
    data['variables']['imageVersion'] = vhd_version
    data['variables']['imageOffer'] = offer_id
    data['variables']['channel'] = release_channel
    data['variables']['storageAccountDNSSuffix'] = storage_suffix
    data['variables']['portalFQDN'] = portal_endpoint
    data['variables']['editionAddOn'] = edition_addon

    # Use multiple steps to keep order
    parameters = data.get('parameters')
    if parameters:
        new_parameters = {
            "ddcUsername": {
                "defaultValue": "admin",
                "type": "String",
                "metadata": {
                    "description": "Please enter the username you want to use for Docker Enterprise Edition."
                }
            },
            "ddcPassword": {
                "minLength": 8,
                "maxLength": 40,
                "type": "SecureString",
                "metadata": {
                    "description": "Please enter the password you want to use for Docker Enterprise Edition."
                }
            },
            "ddcLicense": {
                "type": "SecureString",
                "metadata": {
                    "description": "Upload your Docker Enterprise Edition License Key"
                }
            }
        }
        parameters.update(new_parameters)
        # Let DDC handle the logs
        del parameters["enableExtLogs"]

    variables = data.get('variables')
    if variables:
        new_variables = {
            "ddcUser": "[parameters('ddcUsername')]",
            "ddcPass": "[parameters('ddcPassword')]",
            "ddcLicense": "[base64(string(parameters('ddcLicense')))]",
            "DockerProviderTag": "8CF0E79C-DF97-4992-9B59-602DB544D354",
            "lbDTRFrontEndIPConfigID": "[concat(variables('lbSSHID'),'/frontendIPConfigurations/dtrlbfrontend')]",
            "lbDTRName": "dtrLoadBalancer",
            "ucpTag": UCP_TAG,
            "dtrTag": DTR_TAG,
            "lbDTRPublicIPAddressName": "[concat(variables('basePrefix'), '-', variables('lbDTRName'),  '-public-ip')]",
            "lbPublicIpDnsName": "[concat('applb-', variables('groupName'))]",
            "ucpLbPublicIpDnsName": "[concat('ucplb-', variables('groupName'))]",
            "dtrLbPublicIpDnsName": "[concat('dtrlb-', variables('groupName'))]",
            "ucplbname": "[concat(variables('ucpLbPublicIpDnsName'), '.', variables('storageLocation'), '.', '" + resource_mgr_vm_suffix + "')]",
            "dtrlbname": "[concat(variables('dtrLbPublicIpDnsName'), '.', variables('storageLocation'), '.', '" + resource_mgr_vm_suffix + "')]",
            "lbpublicIPAddress1": "[resourceId('Microsoft.Network/publicIPAddresses',variables('lbSSHPublicIPAddressName'))]",
            "lbpublicIPAddress2": "[resourceId('Microsoft.Network/publicIPAddresses',variables('lbDTRPublicIPAddressName'))]",
            "dtrStorageAccount": "[concat(uniqueString(concat(resourceGroup().id, variables('storageAccountSuffix'))), 'dtr')]"
        }
        variables.update(new_variables)
        variables["lbSSHFrontEndIPConfigID"] = "[concat(variables('lbSSHID'),'/frontendIPConfigurations/ucplbfrontend')]"
        variables["subnetPrefix"] = "172.31.0.0/24"
    resources = data.get('resources')
    # Add DTR Storage Account
    dtr_resources = [
        {
            "type": "Microsoft.Storage/storageAccounts",
            "name": "[variables('dtrStorageAccount')]",
            "apiVersion": "variables('storApiVersion')]",
            "location": "[variables('storageLocation')]",
            "tags": {
                "provider": "[toUpper(variables('DockerProviderTag'))]"
            },
            "kind": "Storage",
            "sku": {
                "name": "Standard_LRS"
            }
        },
        {
            "type": "Microsoft.Network/publicIPAddresses",
            "name": "[variables('lbDTRPublicIPAddressName')]",
            "apiVersion": "[variables('apiVersion')]",
            "location": "[variables('storageLocation')]",
            "tags": {
                "provider": "[toUpper(variables('DockerProviderTag'))]"
            },
            "properties": {
                "publicIPAllocationMethod": "Static",
                "dnsSettings": {
                    "domainNameLabel": "[variables('dtrLbPublicIpDnsName')]"
                }
            }
        }
    ]
    resources.extend(dtr_resources)
    for key, val in enumerate(resources):
        #  Add Docker Provider tag to all resources
        tag_rule = {
            "tags": {
                "provider": "[toUpper(variables('DockerProviderTag'))]",
                "channel": "[variables('channel')]"
            }
        }
        val.update(tag_rule)
        if 'properties' not in val:
            continue
        properties = val['properties']
        if val['name'] == "[variables('managerNSGName')]":
            new_security_rule = [
                {
                    "name": "ucp",
                    "properties": {
                        "access": "Allow",
                        "description": "Allow UCP",
                        "destinationAddressPrefix": "*",
                        "destinationPortRange": "12390",
                        "direction": "Inbound",
                        "priority": 207,
                        "protocol": "Tcp",
                        "sourceAddressPrefix": "*",
                        "sourcePortRange": "*"
                    }
                },
                {
                    "name": "dtr443",
                    "properties": {
                        "access": "Allow",
                        "description": "Allow DTR",
                        "destinationAddressPrefix": "*",
                        "destinationPortRange": "12391",
                        "direction": "Inbound",
                        "priority": 208,
                        "protocol": "Tcp",
                        "sourceAddressPrefix": "*",
                        "sourcePortRange": "*"
                    }
                },
                {
                    "name": "dtr80",
                    "properties": {
                        "access": "Allow",
                        "description": "Allow DTR",
                        "destinationAddressPrefix": "*",
                        "destinationPortRange": "12392",
                        "direction": "Inbound",
                        "priority": 209,
                        "protocol": "Tcp",
                        "sourceAddressPrefix": "*",
                        "sourcePortRange": "*"
                    }
                }
            ]
            properties['securityRules'].extend(new_security_rule)
        if val['name'] == "[variables('lbPublicIPAddressName')]":
            new_dns_property = {
                "dnsSettings": {
                    "domainNameLabel": "[variables('lbPublicIpDnsName')]"
                }
            }
            properties.update(new_dns_property)
        if val['name'] == "[variables('lbSSHPublicIPAddressName')]":
            new_dns_property = {
                "dnsSettings": {
                    "domainNameLabel": "[variables('ucpLbPublicIpDnsName')]"
                }
            }
            properties.update(new_dns_property)
        if val['name'] == "[variables('lbSSHName')]":
            properties["frontendIPConfigurations"] = [
                    {
                        "name": "ucplbfrontend",
                        "properties": {
                            "publicIPAddress": {
                                "id": "[variables('lbpublicIPAddress1')]"
                            }
                        }
                    },
                    {
                        "name": "dtrlbfrontend",
                        "properties": {
                            "publicIPAddress": {
                                "id": "[variables('lbpublicIPAddress2')]"
                            }
                        }
                    }
                ]
            loadbalancing_rules = {
                "loadBalancingRules": [
                    {
                        "name": "ucpLbRule",
                        "properties": {
                            "backendPort": 12390,
                            "enableFloatingIP": False,
                            "frontendIPConfiguration": {
                                "id": "[variables('lbSSHFrontEndIPConfigID')]"
                            },
                            "backendAddressPool": {
                                "id": "[concat(variables('lbSSHID'), '/backendAddressPools/default')]"
                            },
                            "frontendPort": 443,
                            "idleTimeoutInMinutes": 5,
                            "probe": {
                                "id": "[concat(variables('lbSSHID'),'/probes/ucp443')]"
                            },
                            "protocol": "tcp"
                        }
                    },
                    {
                        "name": "dtrLbRuleHTTPS",
                        "properties": {
                            "backendPort": 12391,
                            "enableFloatingIP": False,
                            "frontendIPConfiguration": {
                                "id": "[variables('lbDTRFrontEndIPConfigID')]"
                            },
                            "backendAddressPool": {
                                "id": "[concat(variables('lbSSHID'), '/backendAddressPools/default')]"
                            },
                            "frontendPort": 443,
                            "idleTimeoutInMinutes": 5,
                            "probe": {
                                "id": "[concat(variables('lbSSHID'),'/probes/dtr443')]"
                            },
                            "protocol": "tcp"
                        }
                    },
                    {
                        "name": "dtrLbRuleHTTP",
                        "properties": {
                            "backendPort": 12392,
                            "enableFloatingIP": False,
                            "frontendIPConfiguration": {
                                "id": "[variables('lbDTRFrontEndIPConfigID')]"
                            },
                            "backendAddressPool": {
                                "id": "[concat(variables('lbSSHID'), '/backendAddressPools/default')]"
                            },
                            "frontendPort": 80,
                            "idleTimeoutInMinutes": 5,
                            "probe": {
                                "id": "[concat(variables('lbSSHID'),'/probes/dtr80')]"
                            },
                            "protocol": "tcp"
                        }
                    }
                ]
            }
            properties.update(loadbalancing_rules)
            new_probe = [
                {
                    "name": "ucp443",
                    "properties": {
                        "intervalInSeconds": 10,
                        "numberOfProbes": 2,
                        "port": 12390,
                        "protocol": "tcp"
                    }
                },
                {
                    "name": "dtr443",
                    "properties": {
                        "intervalInSeconds": 10,
                        "numberOfProbes": 2,
                        "port": 12391,
                        "protocol": "tcp"
                    }
                },
                {
                    "name": "dtr80",
                    "properties": {
                        "intervalInSeconds": 10,
                        "numberOfProbes": 2,
                        "port": 12392,
                        "protocol": "tcp"
                    }
                }
            ]
            properties['probes'].extend(new_probe)
            new_depends = [
                "[concat('Microsoft.Network/publicIPAddresses/', variables('lbDTRPublicIPAddressName'))]"
            ]
            val['dependsOn'].extend(new_depends)
    outputs = data.get('outputs')
    if outputs:
        new_outputs = {
            "UCPLoginURL": {
                "type": "String",
                "value": "[concat('https://', reference(resourceId('Microsoft.Network/publicIPAddresses', variables('lbSSHPublicIPAddressName'))).dnsSettings.fqdn)]"
            },
            "DTRLoginURL": {
                "type": "String",
                "value": "[concat('https://', reference(resourceId('Microsoft.Network/publicIPAddresses', variables('lbDTRPublicIPAddressName'))).dnsSettings.fqdn)]"
            }
        }
        outputs.update(new_outputs)

    outdir = u"dist/azure/{}/{}".format(release_channel, docker_for_azure_version)
    # if the directory doesn't exist, create it.
    if not os.path.exists(outdir):
        os.makedirs(outdir)

    outfile = u"{}/{}.tmpl".format(outdir, arm_template_name)

    with open(outfile, 'w') as outf:
        json.dump(data, outf, indent=4, sort_keys=True)

    # TODO: json validate
    if is_json(outfile):
      print(u"ARM template created in {}".format(outfile))
    else:
      print(u"ERROR: ARM template invalid in {}".format(outfile))

    return outfile

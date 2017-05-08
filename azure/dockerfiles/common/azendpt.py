#!/usr/bin/env python

# Endpoint for Azure platforms like China, Germany, etc. Refer to:
# https://github.com/Azure/go-autorest/blob/master/autorest/azure/environments.go
# For on-prem Azure stack, we will need to get these as parameters from admin

AZURE_PLATFORMS = {
    "PUBLIC" : {
        "PORTAL_ENDPOINT"                   : u"portal.azure.com",
        "TEMPLATE_SUFFIX"                   : "",
        "PUBLIC_PLATFORM"                   : True,
        "STORAGE_ENDPOINT"                  : u"core.windows.net",
        "STORAGE_BLOB_SUFFIX"               : u".blob.core.windows.net",
        "RESOURCE_MANAGER_ENDPOINT"         : u"https://management.azure.com/",
        "ACTIVE_DIRECTORY_ENDPOINT"         : u"https://login.microsoftonline.com/",
        "SERVICE_MANAGEMENT_ENDPOINT"       : u"https://management.core.windows.net/"
    },
    "GOV" : {
        "PORTAL_ENDPOINT"                   : u"portal.azure.us",
        "TEMPLATE_SUFFIX"                   : u"-gov",
        "PUBLIC_PLATFORM"                   : False,
        "STORAGE_ENDPOINT"                  : u"core.usgovcloudapi.net",
        "STORAGE_BLOB_SUFFIX"               : u".blob.core.usgovcloudapi.net",
        "RESOURCE_MANAGER_ENDPOINT"         : u"https://management.usgovcloudapi.net/",
        "ACTIVE_DIRECTORY_ENDPOINT"         : u"https://login.microsoftonline.com/",
        "SERVICE_MANAGEMENT_ENDPOINT"       : u"https://management.core.usgovcloudapi.net/"
    }
}

AZURE_DEFAULT_ENV = "PUBLIC"

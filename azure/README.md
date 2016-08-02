# Docker for Azure

## Development

Current development instructions:

- Must have Azure X-Platform CLI (will containerize this eventually) and be
  authenticated via ARM model
- Must be using Editions Azure account (due to `editionsstorage` storage account
  where Moby VHD is maintained)
- Must have GNU make

If these are true, run:

```
$ make
```

and you will create a resource group (name in `groupname.out`) from the
`editions.json`.  It should prompt you for your SSH public key.

## Azure VHD Creation and Publishing Process

To create and publish a VHD requires several steps.

1. Build local Moby VHD 
2. Upload Moby VHD to an Azure storage account (similar to S3 bucket)
3. Create a SAS URL using Azure Storage Explorer so that the Azure publishing
   portal can obtain the uploaded VHD
4. Submit the image in the publishing portal and "push to staging"
5. (Update `editions.json` ARM template?)

#### Building and Uploading

(Note: The following section will require an Azure account.  If you do not have
one please talk to Olivier Gambier or Madhu Venugopal to obtain access.)

Start in `alpine/` directory of the [Moby
repository](https://github.com/docker/moby).

Obtain a Azure storage account key for accessing the blob storage container
where the VHD will be stored.  This is available under the little "key" icon in
the portal. Set this value to `AZURE_STG_ACCOUNT_KEY` in your shell.

```
export AZURE_STG_ACCOUNT_KEY="biglongstring"
```

Once that is set, `make azure` will generate and upload the Moby VHD for you
without needing any dependencies besides Docker (all of the build happens in
Docker containers). You may want to `sudo make clean` first to ensure the build
is pristine.

__NOTE__: There is an issue with the Azure program used to upload the artifact
where it will sometimes hang and/or panic.  If this happens try running `sudo
make clean` and `make azure` again.  It will likely succeed with multiple
attempts.

```
$ sudo make clean
$ make azure
```

At the end of the build you will see output indicating where the VHD was
uploaded.

```console
....
Upload completed
 ---> VHD uploaded.
 ---> https://dockereditions.blob.core.windows.net/mobylinux/6d2b2b7d3b466b6f54737d6e0076ea6d-mobylinux.vhd
```

Take note of the VHD's name and location inside the storage container.

#### Create SAS URL

In order for the Azure folks to upload your image to the marketplace they need
you to provide a SAS ("Shared Access Signature") URL for the blob so that they
can pull it down for testing.  A SAS URL is a temporary use URL that will grant
read or write access to otherwise private resources in a storage container.

To create one, download the [Azure Storage
Explorer](https://azurestorageexplorer.codeplex.com/) program for Windows.
Connect it to the relevant storage account and navigate to the container for
your uploaded VHD.  Find the VHD file in the list, highlight it, and click on
"Security" icon in the left hand side panel.

A little popup for "Blob Container Security" will come up.  Click on the "Shared
Access Signatures" tab.

In this tab:

- The default one-week access period is fine to leave as-is
- Select only "Read" and "List" checkboxes
- Click "Generate Signature"
- Change the `http://` in the generated signature to `https://` to ensure the
  artifact is distributed using HTTPS protocol (HTTP is insecure)
- Copy the generated SAS URI to the clipboard

#### Submit and Push to Staging in Portal

In the [VM images
tab](https://publish.windowsazure.com/workspace/Offers/AzureStore/Docker-forAzureImage-d29b/vm-images)
for the Docker for Azure image update the existing "VIRTUAL MACHINE IMAGES"
section or create a new image version.  The "OS VHD URL" is where you will place
the generated SAS URL.  After pasting this SAS URL click "Save" in the upper
right hand corner of this page.

Once you have updated or created a new VM image click on the "Publish" link
nested under "Docker for Azure" image in the left hand side.  Click the "PUSH TO
STAGING" button to stage the VHD image.  A popup will prompt you to specify
which accounts should be allowed to have access to the image.  This will take a
bit of time as they need to run through checks to ensure the image fulfills
their Marketplace requirements.

#### Update `editions.json` ARM template

This is not yet implemented today but the final step will be updating the Docker
for Azure ARM template to indicate that the new Marketplace image should be
used.

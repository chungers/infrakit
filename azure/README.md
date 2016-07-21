# Docker for Azure

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

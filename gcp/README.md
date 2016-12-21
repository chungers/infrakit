# Docker for Google Cloud Platform

## Installation from Google Cloud Shell

1. Open Cloud Shell for a GCP project
2. Run `curl -sSL http://get.docker-gcp.com/ | sh`
3. There's no step 3

## Installation from the command line

### Prerequisites

- Access to a Google Cloud project with those Apis enabled:
  - [Google Cloud Deployment Manager V2 API](https://console.developers.google.com/apis/api/deploymentmanager-json.googleapis.com/overview?project=docker4x&duration=PT1H)
  - [Google Cloud RuntimeConfig API](https://console.developers.google.com/apis/api/runtimeconfig.googleapis.com/overview?project=docker4x)
- Have `make` installed

### Create a swarm

```
make auth
make create
```

### Delete a swarm

```
make delete
make revoke
```

# Docker for Google Cloud Platform

## Installation from Google Cloud Shell

1. Open Cloud Shell for a GCP project
2. Run `curl -sSL http://get.docker-gcp.com/ | sh`
3. There's no step 3

## Installation from the command line

### Prerequisites

- Have `make` installed
- Access to a Google Cloud project
- Enable [Google Cloud Deployment Manager V2 API](https://console.developers.google.com/apis/api/deploymentmanager-json.googleapis.com/overview?project=docker4x&duration=PT1H)
- Enable [Google Cloud RuntimeConfig API](https://console.developers.google.com/apis/api/runtimeconfig.googleapis.com/overview?project=docker4x)

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

# TODO

 + External load balancer
 + Better UX
 + Use Moby
 + Monitoring
 + SSH keys
 + Additional swarm properties
 + DTR/DDC
 + Diagnostics

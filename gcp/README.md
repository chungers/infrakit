# Docker for Google Cloud Platform

## Prerequisites

- Access to a Google Cloud project
- Enable [Google Cloud Deployment Manager V2 API](https://console.developers.google.com/apis/api/deploymentmanager-json.googleapis.com/overview?project=docker4x&duration=PT1H)
- Enable [Google Cloud RuntimeConfig API](https://console.developers.google.com/apis/api/runtimeconfig.googleapis.com/overview?project=docker4x)

## Create a swarm

```
make auth
make create
```

## Delete a swarm

```
make delete
make revoke
```

# TODO

 + External load balancer
 + Better UX
 + Use Moby
 + Use Google Log driver
 + Monitoring
 + Configure project
 + SSH keys
 + Additional swarm properties
 + Publish the templates
 + See how the Cloud Shell fits in the big picture
 + DTR/DDC
 + Auto-enable all the Apis we need
 + Have each worker increment a counter to be able to wait from outside
 + Diagnostics

# Docker for Google Cloud Platform

## Prerequisites

- Access to a Google Cloud account.
- Enable some Apis (TODO: I have yet to document this list!)
- Install the `gcloud` command line
- Run `gcloud auth login`

## Create a swarm

```
make create
```

## Delete a swarm

```
make delete
```

# TODO

 + External load balancer
 + Better UX
 + Use Moby
 + Logs
 + Monitoring
 + Configure project
 + Multiple managers
 + Different machine types for workers and managers
 + SSH keys
 + Additional swarm properties
 + Publish the templates
 + See how the Cloud Shell fits in the big picture
 + DTR/DDC
 + List the Apis we need
 + Auto-enable all the Apis we need

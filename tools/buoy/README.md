# Buoy
This is the go binary that will send analytic information to segment for Docker for AWS

## Usage

### Identify the swarm
This is only done once on swarm init, to register the swarm. It will get it's variables from
the ENV (stack_id, account_id, and region)

$ buoy -identify

### Ping
We want to ping every hour or so, this will let us know the status of the cluster

# this will ping, and tell us that it has 5 worker nodes, 3 manager nodes, 7 services and the docker_version is 1.12.0-rc3
# it will also pass in the docker for aws version via an ENV variable.
$ buoy -workers=5 -managers=3 -services=7 -docker_version=1.12.0-rc3

## Building
Run `./build_buoy.sh` it will build the Go binary in a docker container and put the results in the `aws/dockerfiles/files/bin/`` directory.

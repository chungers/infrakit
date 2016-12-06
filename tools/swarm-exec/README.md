# swarm-exec
A simple tool to execute a docker command in all the swarm nodes. It uses swarm's global service mode to execute the command in all the nodes and bind-mounts docker cli and docker socket.

## Releasing
`./release.sh`

## Building
`./build.sh`


## Requirements
- This needs to run on a swarm manager, and it needs access to the Docker command-line in order to run the commands.
- Docker 1.12 or higher needs to be running on all of the hosts on your swarm

### docker image
Run the image passing in the different commands.

docker run -it -v /usr/bin/docker:/usr/bin/docker -v /var/run/docker.sock:/var/run/docker.sock docker4x/swarm-exec:v0.1 swarm-exec docker version

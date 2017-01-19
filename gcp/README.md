# Docker for GCP

## Prerequisites

- Access to a Google Cloud project with those Api enabled:
  - [Google Cloud Deployment Manager V2 API](https://console.developers.google.com/apis/api/deploymentmanager-json.googleapis.com/overview?project=docker4x&duration=PT1H)
  - [Google Cloud RuntimeConfig API](https://console.developers.google.com/apis/api/runtimeconfig.googleapis.com/overview?project=docker4x)
- Make sure that you have enough capacity for the swarm that you want to build, and won't go over any of your limits.

Once you have all of the above you are ready to move onto the next step.

## Configuration

You can use the `gcloud` CLI to invoke the template.  e.g.:

    $ gcloud init --skip-diagnostics
    $ gcloud deployment-manager deployments create [name_of_the_stack] \
      --config https://storage.googleapis.com/docker-for-gcp-templates/gcp-v1.13.0-rc6-beta16/swarm.jinja \
      --properties managerCount:3,workerCount:2,managerMachineType:g1-small,workerMachineType:g1-small

Note: You don't need to install `gcloud` nor do you have to `gcloud init` if you run
the `create` command on the Cloud Shell for your GCP project.

Note: The `Makefile` in this directory in the repository can invoke this for you
automatically:

    $ make auth
    $ make create

It can also tear down created stack(s) via `make clean`, e.g.:

    $ make delete
    $ make revoke

You can choose how many managers and workers you want to run depending on how
much resilience you need.

    | # of managers  | # of tolerated failures |
    | ------------- | ------------- |
    | 1  | 0  |
    | 3  | 1  |
    | 5  | 2  |

Once the stack is deployed, it will print the Swarm's external IP address and
the command line to connect to a manager. It's something like:

    $ gcloud compute ssh --zone europe-west1-d docker-manager-1

Once you are logged into the node you can run Docker commands on the cluster:

    $ docker swarm info
    $ docker node ls

You can also tunnel the Docker socket over SSH to remotely run commands on the cluster:

    $ gcloud compute ssh --zone europe-west1-d docker-manager-1 -- -NL localhost:2375:/var/run/docker.sock &
    $ docker -H localhost:2375 info

## Running apps

You can now start creating containers and services.

    $ docker run hello-world

To run websites, you to expose a port:

    $ docker network create -d overlay foo
    $ docker service create --name nginx --scale 1 --network foo -p 80:80/tcp nginx

Once up, you access the site on the Swarm's external IP address.

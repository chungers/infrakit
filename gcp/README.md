# Docker for GCP

## Prerequisites

- Access to a Google Cloud project with those Api enabled:
  - [Google Cloud Deployment Manager V2 API](https://console.developers.google.com/apis/api/deploymentmanager-json.googleapis.com/overview?project=docker4x&duration=PT1H)
  - [Google Cloud RuntimeConfig API](https://console.developers.google.com/apis/api/runtimeconfig.googleapis.com/overview?project=docker4x)
- Make sure that you have enough capacity for the swarm that you want to build, and won't go over any of your limits.
- Install the [Cloud SDK](https://cloud.google.com/sdk/downloads) (`gcloud`). It's not a hard prerequisite but makes interacting with your GCP project easier.

Once you have all of the above you are ready to move onto the next step.

## Configuration

To create a stack called `docker`:

    $ make auth
    $ make create

It can also tear down created stack(s):

    $ make delete
    $ make revoke

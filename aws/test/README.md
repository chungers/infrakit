# AWS End to End Testing

### Overview
The Makefile executes a full end to end test, and can be incorporated with a Jenkins build.
This process involves the following steps:

1. Obtain the template that was created from the Jenkins build
2. Log into aws, and create a stack based on the template
3. SSH into a manager and run a set of tests

The main driver of the stack deployment is the createStack.sh script.
The main driver of the tests is the [e2e tests](https://github.com/docker/docker-e2e)


### Additional Notes
The aws commands are run in a docker container (`docker4x/awscli`)
The end to end tests are also run from within a container.

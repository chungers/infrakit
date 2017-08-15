# Regression Test Framework (RTF)

### Overview
To test Docker for Azure, and Docker for AWS we are harnessing [rtf](https://github.com/linuxkit/rtf).
rtf is a simple shell based test framework that can be run within a docker container. To run locally you can
also use `go get -u github.com/linuxkit/rtf` to install rtf. A set of utility functions exist in `/cases/lib.sh` in
order to facilitate test writing.

Using rtf allows us to test expected docker functionality as well as the Docker CLI simultaneously.
Having the test cases written directly in the Docker CLI also makes it easier to understand the tests, and 
allows for manual reproducibility if necessary.

For more information on the design of rtf refer to [DESIGN.md](https://github.com/linuxkit/rtf/blob/master/docs/DESIGN.md).
For a full user guide refer to [USER_GUIDE.md](https://github.com/linuxkit/rtf/blob/master/docs/USER_GUIDE.md)
The user guide provides more information about using labels for fine grain run control, describes general test writing guidelines, as well as additional features. 

### Quick Run
All the test will be in an easy to use container. To run the tests simply run:
```
docker container run -v /usr/bin/docker:/usr/bin/docker -v /var/run/docker.sock:/var/run/docker.sock corbincoleman/rtf run
```
All the rtf flags are available to use as well. As an example for more verbose output you can run the container like this:
```
docker container run -v /usr/bin/docker:/usr/bin/docker -v /var/run/docker.sock:/var/run/docker.sock corbincoleman/rtf -vvv run
```

### Deployment
The set of test cases will be deployed from within one of the corresponding end to end test frameworks from within a container 
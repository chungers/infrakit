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
The tests only need to have a `DOCKER_VERSION` specified for them to run properly. It can be specified when using `make` to run the container. If no `DOCKER_VERSION` is specified it will default to using `17.06.1-ce`.
```
make run DOCKER_VERSION='your docker version'
```
Making sure that `DOCKER_VERSION` is specified only effects the `expected_version` test. All other tests will run properly if this is left out from the `make` command. Adding the `DOCKER_VERSION` to the `make` command will use the variable as a `build ARG` when building the container. The `build ARG` is saved in the container in an environment variable called `EXPECTED_VERSION`.

The tests can also be run with more verbose output. This will display all commands, stdout, and stderr messages ran/displayed while running each of the tests.
```
make run-verbose DOCKER_VERSION='your docker version'
```
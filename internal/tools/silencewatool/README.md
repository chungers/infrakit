A tool primarily to be used by support to flip the WALinuxAgent config to non verbose mode and restart WALinuxAgent.

BUILD:

docker build -t docker4x/silence-walinux -f Dockerfile .

EXECUTE:

swarm-exec docker run -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker --log-driver=json-file docker4x/silence-walinux

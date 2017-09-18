FROM alpine:3.5

MAINTAINER Ken Cochrane <Ken.Cochrane@Docker.com>

RUN apk add --update bash ca-certificates jq groff less python py-pip py-setuptools \
    && pip install awscli \
    && pip install https://s3.amazonaws.com/cloudformation-examples/aws-cfn-bootstrap-latest.tar.gz \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

# Add the docker client to the container (don't rely on host binary)
ARG DOCKER_BIN_URL=http://docker-ci-artifacts.s3.amazonaws.com/ci.qa.aws.dckr.io/jenkins-docker-release-packaging-ce-17.06-12/docker-amd64.tgz

ADD $DOCKER_BIN_URL /docker-amd64.tgz
RUN tar -zxvf /docker-amd64.tgz --strip-components 1 -C /usr/bin docker/docker
RUN rm /docker-amd64.tgz

# setup init
RUN mkdir -p /usr/bin/ /var/log/docker/
COPY files/entry.sh /entry.sh
ADD bin/buoy /usr/bin
RUN chmod 755 /entry.sh /usr/bin/buoy

CMD ["/entry.sh"]

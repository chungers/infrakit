FROM alpine:3.5

MAINTAINER KenCochrane <Ken.Cochrane@Docker.com>

RUN apk add --update bash ca-certificates jq groff python py-pip py-setuptools curl \
    && pip install awscli \
    && pip install https://s3.amazonaws.com/cloudformation-examples/aws-cfn-bootstrap-latest.tar.gz \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

# Add the docker client to the container (don't rely on host binary)
ARG DOCKER_BIN_URL=http://docker-ci-artifacts.s3.amazonaws.com/ci.qa.aws.dckr.io/jenkins-docker-release-packaging-ce-17.06-12/docker-amd64.tgz

ADD $DOCKER_BIN_URL /docker-amd64.tgz
RUN tar -zxvf /docker-amd64.tgz --strip-components 1 -C /usr/bin docker/docker
RUN rm /docker-amd64.tgz

# setup cron
RUN mkdir -p /usr/docker /var/log/docker/
ADD files/crontab.txt /usr/docker/crontab.txt
ADD files/watcher.sh /usr/bin
ADD files/cleanup.sh /usr/bin
ADD files/buoy.sh /usr/bin
ADD files/refresh.sh /usr/bin
ADD files/vacuum.sh /usr/bin
ADD files/bouncer.sh /usr/bin
ADD bin/buoy /usr/bin
COPY files/entry.sh /entry.sh
RUN chmod 755 /usr/bin/watcher.sh /entry.sh /usr/bin/cleanup.sh \
    /usr/bin/buoy /usr/bin/buoy.sh /usr/bin/vacuum.sh /usr/bin/refresh.sh \
    /usr/bin/bouncer.sh
RUN /usr/bin/crontab /usr/docker/crontab.txt

CMD ["/entry.sh"]

FROM alpine:3.5

MAINTAINER KenCochrane <Ken.Cochrane@Docker.com>

RUN apk add --update bash ca-certificates jq groff python py-pip py-setuptools curl \
    && pip install awscli \
    && pip install https://s3.amazonaws.com/cloudformation-examples/aws-cfn-bootstrap-latest.tar.gz \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

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

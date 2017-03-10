FROM alpine:3.5

RUN apk add --update ca-certificates curl jq

CMD ["/entry.sh"]

RUN mkdir -p /usr/docker && mkdir -p /var/log/docker/
COPY files/crontab.txt /usr/docker/
COPY files/vacuum.sh files/cleanup.sh files/buoy.sh bin/buoy /usr/bin/
COPY files/entry.sh /entry.sh
RUN /usr/bin/crontab /usr/docker/crontab.txt

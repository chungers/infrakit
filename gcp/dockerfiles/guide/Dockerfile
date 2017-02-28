FROM alpine:3.5

RUN apk add --update ca-certificates curl jq

# needed in order for go binary to work.
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

CMD ["/entry.sh"]

RUN mkdir -p /usr/docker && mkdir -p /var/log/docker/
COPY files/crontab.txt /usr/docker/
COPY files/vacuum.sh files/cleanup.sh files/buoy.sh bin/buoy /usr/bin/
COPY files/entry.sh /entry.sh
RUN /usr/bin/crontab /usr/docker/crontab.txt

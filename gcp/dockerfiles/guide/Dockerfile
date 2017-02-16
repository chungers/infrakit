FROM alpine:3.5

RUN apk add --update ca-certificates jq

# needed in order for go binary to work.
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

CMD ["/entry.sh"]

RUN mkdir -p /usr/docker /var/log/docker/
COPY crontab.txt /usr/docker/
COPY entry.sh /
COPY vacuum.sh cleanup.sh buoy.sh bin/buoy /usr/bin/
RUN /usr/bin/crontab /usr/docker/crontab.txt

FROM alpine:3.5

RUN apk add --update bash openssh ca-certificates sudo curl jq tar python py-setuptools shadow

ENTRYPOINT ["/entry.sh"]

RUN mkdir -p /usr/docker /var/log/docker/
COPY crontab.txt /usr/docker/
COPY entry.sh /
COPY vacuum.sh cleanup.sh buoy.sh /usr/bin/
COPY bin/buoy /usr/bin

RUN /usr/bin/crontab /usr/docker/crontab.txt

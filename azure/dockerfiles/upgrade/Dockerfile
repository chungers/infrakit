FROM alpine:3.5

MAINTAINER Deep Debroy <ddebroy@docker.com>

RUN apk add --update bash ca-certificates curl \
    && rm -rf /var/cache/apk/*

ARG CHANNEL
ENV DOCKER_CHANNEL=$CHANNEL

ARG VERSION
ENV DOCKER_VERSION=$VERSION

# setup init
RUN mkdir -p /var/log/docker/
ADD files/upgrade.sh /usr/bin/upgrade.sh
ADD files/azupgrade.py /opt/azupgrade.py
ADD files/azupgrade_log_cfg.json /opt/azupgrade_log_cfg.json
RUN chmod 755 /usr/bin/upgrade.sh /opt/azupgrade.py
CMD ["/usr/bin/upgrade.sh"]

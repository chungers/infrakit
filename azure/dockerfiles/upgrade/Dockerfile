FROM alpine:3.5

MAINTAINER Deep Debroy <ddebroy@docker.com>

RUN apk add --update bash ca-certificates jq groff less openssl-dev libffi-dev python python-dev musl-dev linux-headers py-pip gcc curl py-setuptools \
    && pip install msrest \
    && pip install msrestazure \
    && pip install --pre azure \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

ARG VERSION
ENV DOCKER_VERSION=$VERSION

# setup init
RUN mkdir -p /var/log/docker/
ADD files/upgrade.sh /usr/bin/upgrade.sh
ADD files/azupgrade.py /opt/azupgrade.py
ADD files/azupgrade_log_cfg.json /opt/azupgrade_log_cfg.json
ADD files/azendpt.py /opt/azendpt.py
ADD files/dockerutils.py /opt/dockerutils.py
ADD files/aztags.py /opt/aztags.py
RUN chmod 755 /usr/bin/upgrade.sh /opt/azupgrade.py  /opt/aztags.py
CMD ["/usr/bin/upgrade.sh"]

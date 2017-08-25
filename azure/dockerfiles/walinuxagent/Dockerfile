FROM docker4x/shell-aws:compose@sha256:5e851a3ed38aa2c67244fcfafa4b4d9fae035bf73e8b7384f4c4fecc8e42cb64
# Created from aws shell tag

RUN apk add --no-cache --update \
        sfdisk \
        supervisor \
        gawk \
        rsyslog \
        openssl \
        ca-certificates \
        openssh \
        parted \
        sudo \
        net-tools \
        ifupdown \
        python \
        util-linux \
        git \
        py-setuptools \
        py2-pip

# Windows Azure Linux agent section
RUN pip install pyasn1
ENV AGENT_REVISION 5d66b1c923fcccf1f00b6004627f7f3d46172142

# TODO: Move back to upstream (a PR is out for this on the Azure agent --
# https://github.com/Azure/WALinuxAgent/pull/341), for now we have to use this
# fork because detection doesn't work in upstream
RUN git clone https://github.com/Azure/WALinuxAgent /WALinuxAgent
WORKDIR /WALinuxAgent
RUN git reset --hard ${AGENT_REVISION}
RUN python setup.py install
COPY waagent.conf /opt/waagent.conf
RUN cp bin/* /usr/sbin/

# TODO: If upstream shell Dockerfile changes this will also need to be changed.
# Should be implemented more elegantly.
RUN mkdir /daemons && \
    rm /entry.sh && \
    mv /etc/ssl/openssl.cnf /opt/openssl.cnf && \
    mv /etc/ssh/sshd_config /opt/sshd_config && \
    mv /etc/motd /opt/motd && \
    cp /etc/logrotate.d/waagent.logrotate /opt/waagent.logrotate

COPY /walinuxagent-entrypoint.sh /daemons/walinuxagent-entrypoint.sh
COPY /supervisord.conf /opt/supervisord.conf
COPY /upgrade.sh /usr/bin/upgrade.sh

RUN chmod +x /usr/sbin/waagent && \
    ln -sf /dev/stdout /var/log/waagent.log
RUN chmod +x /usr/bin/upgrade.sh

ENTRYPOINT ["supervisord"]
CMD ["--configuration", "/opt/supervisord.conf"]

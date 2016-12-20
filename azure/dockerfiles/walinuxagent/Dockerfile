FROM docker4x/shell-aws@sha256:976a48261805160a3bdd834847c44af6097cb3886f5fb80e1c0e459c0c82544e
# Created from aws-v1.14.0-dev-nightly_12_19_2016 tag

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
        py-pip \
        util-linux \
        git \
        py-setuptools

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
    mv /etc/motd /opt/motd

COPY /walinuxagent-entrypoint.sh /daemons/walinuxagent-entrypoint.sh
COPY /upgrade.sh /usr/bin/upgrade.sh
COPY /supervisord.conf /opt/supervisord.conf

RUN chmod +x /usr/bin/upgrade.sh
RUN chmod +x /usr/sbin/waagent && \
    ln -sf /dev/stdout /var/log/waagent.log

ENTRYPOINT ["supervisord"]
CMD ["--configuration", "/opt/supervisord.conf"]

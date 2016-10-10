FROM docker4x/shell

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
ENV AGENT_REVISION ef3d68857cf08b65bbe7dfeb7a811cc9f1aa5bc1

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
    mv /entry.sh /daemons/sshd-entrypoint.sh && \
    mv /etc/ssh/sshd_config /opt/sshd_config && \
    mv /etc/motd /opt/motd

COPY /walinuxagent-entrypoint.sh /daemons/walinuxagent-entrypoint.sh
COPY /supervisord.conf /opt/supervisord.conf

RUN chmod +x /usr/sbin/waagent && \
    ln -sf /dev/stdout /var/log/waagent.log

ENTRYPOINT ["supervisord"]
CMD ["--configuration", "/opt/supervisord.conf"]

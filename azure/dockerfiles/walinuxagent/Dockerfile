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
ENV AGENT_REVISION 78bf8336d57d8349d78ba75cabf224930479fce6

# TODO: Move back to upstream (a PR is out for this on the Azure agent --
# https://github.com/Azure/WALinuxAgent/pull/341), for now we have to use this
# fork because detection doesn't work in upstream
RUN git clone -b supported_dists https://github.com/nathanleclaire/WALinuxAgent /WALinuxAgent
WORKDIR /WALinuxAgent
RUN git reset --hard ${AGENT_REVISION}
RUN python setup.py install
COPY waagent.conf /opt/waagent.conf
RUN cp bin/* /usr/sbin/

# TODO: If upstream shell Dockerfile changes this will also need to be changed.
# Should be implemented more elegantly.
RUN mkdir /daemons && \
    mv /entry.sh /daemons/sshd-entrypoint.sh && \
    mv /etc/ssh/sshd_config /opt/sshd_config

COPY /walinuxagent-entrypoint.sh /daemons/walinuxagent-entrypoint.sh
COPY /supervisord.conf /opt/supervisord.conf

RUN chmod +x /usr/sbin/waagent && \
    ln -sf /dev/stdout /var/log/waagent.log

ENTRYPOINT ["supervisord"]
CMD ["--configuration", "/opt/supervisord.conf"]

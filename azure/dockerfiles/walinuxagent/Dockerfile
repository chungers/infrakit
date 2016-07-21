FROM alpine

RUN apk add --no-cache --update \
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

RUN pip install pyasn1
ENV AGENT_REVISION 54a5b549354e814454112594488230b8905eaa71
RUN git clone -b develop https://github.com/Azure/WALinuxAgent /WALinuxAgent
WORKDIR /WALinuxAgent
RUN git reset --hard ${AGENT_REVISION}
RUN python setup.py install
COPY waagent.conf /opt/waagent.conf
RUN cp bin/* /usr/sbin/
COPY /entrypoint.sh /entrypoint.sh

RUN chmod +x /usr/sbin/waagent && \
    ln -sf /dev/stdout /var/log/waagent.log

ENTRYPOINT ["/entrypoint.sh"]

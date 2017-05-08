FROM alpine:3.5

MAINTAINER Ken Cochrane <Ken.Cochrane@Docker.com>

RUN apk add --update bash ca-certificates jq groff less openssl-dev libffi-dev python python-dev musl-dev linux-headers py-pip gcc curl py-setuptools \
    && pip install msrest \
    && pip install msrestazure \
    && pip install --pre azure \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

# setup init
RUN mkdir -p /usr/bin/ /var/log/docker/
COPY files/entry.sh /entry.sh
COPY files/azureleader.py /usr/bin
COPY files/sakey.py /usr/bin
COPY files/aztags.py /usr/bin
COPY files/azendpt.py /usr/bin
ADD bin/buoy /usr/bin
RUN chmod 755 /entry.sh /usr/bin/azureleader.py /usr/bin/sakey.py \
    /usr/bin/aztags.py /usr/bin/buoy

CMD ["/entry.sh"]

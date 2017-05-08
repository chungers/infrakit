FROM alpine:3.5

MAINTAINER French Ben <frenchben@docker.com>

RUN apk add --no-cache --update bash ca-certificates jq groff less nodejs bash python python-dev linux-headers py-pip py-setuptools make g++ gcc curl libffi-dev openssl-dev\
    && pip install --upgrade pip \
    && npm install -g azure-cli \
    && pip install msrest \
    && pip install msrestazure \
    && pip install --pre azure \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

RUN azure telemetry --disable
RUN azure config mode arm

# setup init
COPY files/entry.sh /entry.sh
COPY files/dtrutils.py /dtrutils.py
COPY files/azendpt.py /azendpt.py
RUN chmod 755 /entry.sh /dtrutils.py


CMD ["/entry.sh"]

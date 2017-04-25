FROM alpine:3.5

MAINTAINER David Chung <david.chung@docker.com>

RUN apk add --no-cache --update --virtual builddeps ca-certificates g++ make python && \
    apk add --no-cache nodejs jq bash && \
    npm install -g azure-cli && \
    apk del builddeps
RUN azure telemetry --disable
RUN azure config mode arm
COPY files/create_sp.sh /usr/bin
RUN chmod 755 /usr/bin/create_sp.sh

ENV PROG docker run -ti docker4x/create-sp-azure

ENTRYPOINT [ "create_sp.sh" ]

FROM alpine:3.5

RUN apk add --update ca-certificates curl jq bash

COPY bin/* /usr/local/bin/
COPY run.sh /run.sh

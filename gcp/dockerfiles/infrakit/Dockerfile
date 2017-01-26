FROM alpine:3.5

RUN apk add --update ca-certificates curl jq bash
WORKDIR /infrakit
CMD ["./run.sh"]

COPY files/* /infrakit/
COPY bin/* /usr/local/bin/

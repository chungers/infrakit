FROM alpine:3.5

MAINTAINER David Chung <david.chung@docker.com>

RUN apk add --update ca-certificates && rm -Rf /tmp/* /var/lib/cache/apk/*

ADD bin/loadbalancer /usr/local/bin/

ENTRYPOINT [ "loadbalancer" ]

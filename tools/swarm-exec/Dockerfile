FROM alpine:3.5
MAINTAINER madhu@docker.com

RUN apk add --update ca-certificates && rm -Rf /tmp/* /var/lib/cache/apk/*

ADD bin/swarm-exec /usr/local/bin/
RUN chmod 755 /usr/local/bin/swarm-exec

CMD [ "swarm-exec" ]

FROM alpine:3.6
MAINTAINER ddebroy@docker.com

RUN apk add --update bash ca-certificates && rm -Rf /tmp/* /var/lib/cache/apk/*

ADD silencewalinux.sh /silencewalinux.sh
RUN chmod 755 /silencewalinux.sh

CMD [ "/silencewalinux.sh" ]

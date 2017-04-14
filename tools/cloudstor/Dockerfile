FROM alpine:3.5

RUN apk add --update bash strace ca-certificates cifs-utils openssl openssl-dev e2fsprogs

RUN rm -rf /usr/share/terminfo
RUN mkdir -p /run/docker/plugins

COPY bin/cloudstor cloudstor

CMD ["cloudstor"]

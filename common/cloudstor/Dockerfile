FROM alpine:3.3

RUN apk add --update bash strace ca-certificates cifs-utils openssl openssl-dev

RUN rm -rf /usr/share/terminfo
RUN mkdir -p /run/docker/plugins

COPY cloudstor cloudstor

CMD ["cloudstor"]

FROM alpine:3.5

RUN apk add --update ca-certificates
ENTRYPOINT ["loadbalancer"]

COPY bin/loadbalancer /usr/local/bin/

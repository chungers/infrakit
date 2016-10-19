FROM alpine:3.3

MAINTAINER KenCochrane <Ken.Cochrane@Docker.com>

RUN apk add --update bash ca-certificates jq groff python py-pip py-setuptools \
    && pip install boto \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

RUN mkdir -p /home/docker /home/docker/dist
ADD tmp/editions.template /home/docker
ADD files/* /home/docker/
RUN chmod 755 /home/docker/*
WORKDIR /home/docker/

ENTRYPOINT ["python", "/home/docker/release.py"]
CMD ["--help"]

FROM alpine:3.5

MAINTAINER Ken Cochrane <Ken.Cochrane@Docker.com>

RUN apk add --update bash ca-certificates jq groff less python py-pip py-setuptools \
    && pip install awscli \
    && pip install https://s3.amazonaws.com/cloudformation-examples/aws-cfn-bootstrap-latest.tar.gz \
    && apk --purge -v del py-pip \
    && rm -rf /var/cache/apk/*

# setup init
RUN mkdir -p /usr/bin/ /var/log/docker/
COPY files/entry.sh /entry.sh
ADD bin/buoy /usr/bin
RUN chmod 755 /entry.sh /usr/bin/buoy

CMD ["/entry.sh"]

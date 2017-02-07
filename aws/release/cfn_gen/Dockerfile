FROM alpine:3.5

MAINTAINER KenCochrane <Ken.Cochrane@Docker.com>

RUN apk add --update bash ca-certificates jq groff python py-pip py-setuptools \
    && pip install awscli boto==2.45.0 \
    && pip install https://s3.amazonaws.com/cloudformation-examples/aws-cfn-bootstrap-latest.tar.gz \
    && rm -rf /var/cache/apk/*

RUN mkdir -p /home/docker
ADD requirements.txt /home/docker/
RUN pip install -r /home/docker/requirements.txt
ADD . /home/docker/
RUN chmod 755 /home/docker/release.sh

CMD ["/home/docker/release.sh"]

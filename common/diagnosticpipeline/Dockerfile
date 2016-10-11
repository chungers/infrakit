FROM alpine

RUN apk add --update --no-cache python3 py-pip coreutils openssl curl
RUN pip install awscli
RUN curl -SsL https://get.docker.com/builds/Linux/x86_64/docker-latest.tgz | tar xvzf - && \
    mv docker/docker /bin/docker && \
    rm -rf docker/ docker.tgz
COPY /bundle-diagnostics.sh /bundle-diagnostics.sh

VOLUME ["/tmp/bucket"]
ENTRYPOINT ["/bundle-diagnostics.sh"]

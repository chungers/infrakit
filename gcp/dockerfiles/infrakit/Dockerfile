FROM alpine:3.5

RUN apk add --update ca-certificates curl jq
RUN mkdir -p /infrakit/plugins /infrakit/configs /infrakit/logs
ENV INFRAKIT_HOME /infrakit
ENV INFRAKIT_PLUGINS_DIR /infrakit/plugins

WORKDIR /infrakit
CMD ["./run.sh"]

COPY files/* /infrakit/
COPY bin/* /usr/local/bin/

FROM alpine

RUN apk add --update --no-cache python3 py-pip
RUN pip install awscli
COPY /bundle-diagnostics.sh /bundle-diagnostics.sh
VOLUME ["/tmp/bucket"]
ENTRYPOINT ["/bundle-diagnostics.sh"]

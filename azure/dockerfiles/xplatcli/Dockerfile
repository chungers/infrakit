FROM alpine

RUN apk add --update --no-cache --virtual builddeps g++ python && \
    apk add --update --no-cache nodejs make && \
    npm install -g azure-cli && \
    apk del builddeps
VOLUME ["/root/.azure"]
ENTRYPOINT ["azure"]

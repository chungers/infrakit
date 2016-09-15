FROM alpine

RUN apk add --update --no-cache --virtual builddeps g++ && \
    apk add --update --no-cache nodejs make python && \
    npm install -g azure-cli && \
    apk del builddeps
VOLUME ["/root/.azure"]
ENTRYPOINT ["azure"]

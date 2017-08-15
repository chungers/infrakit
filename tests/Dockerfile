FROM golang:1.8.3-alpine

RUN mkdir -p /tests/cases
COPY cases /tests/cases

WORKDIR /tests/

RUN apk update && apk upgrade && \
    apk add --no-cache git

RUN go get -u github.com/linuxkit/rtf

CMD ["rtf", "run"]
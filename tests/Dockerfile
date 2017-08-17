FROM golang:1.8.3-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache git 

RUN go get -u github.com/linuxkit/rtf


FROM alpine:latest

COPY --from=0 /go/bin/rtf /usr/local/bin

RUN mkdir -p /tests/cases
COPY cases /tests/cases
WORKDIR /tests/

ENTRYPOINT ["rtf"]
CMD ["help"]


FROM golang:1.7-alpine

RUN apk add --update bash git
RUN go get github.com/LK4D4/vndr

ENV USER root
WORKDIR /go/
COPY . ./
WORKDIR /go/src/metaserver
RUN vndr
WORKDIR /go/

RUN mkdir -p bin

CMD ["./compile.sh"]

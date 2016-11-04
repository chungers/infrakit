FROM golang:1.7

RUN go get github.com/docker/docker/client
RUN go get github.com/aws/aws-sdk-go

ENV USER root
WORKDIR /go/

COPY . ./
RUN mkdir -p bin

CMD ["/compile.sh"]

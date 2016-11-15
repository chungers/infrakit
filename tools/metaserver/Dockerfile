FROM golang:1.7

RUN go get github.com/docker/docker/client
RUN go get github.com/aws/aws-sdk-go
RUN go get github.com/docker/docker/api/types
RUN go get github.com/docker/go-connections/tlsconfig
RUN go get github.com/docker/go-connections/sockets
RUN go get github.com/gorilla/mux

ENV USER root
WORKDIR /go/

COPY . ./
RUN mkdir -p bin

CMD ["/compile.sh"]

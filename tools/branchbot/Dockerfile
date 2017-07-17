FROM golang:1.7-alpine as builder

RUN apk add --update --no-cache git
RUN go get -u github.com/LK4D4/vndr
COPY . /go/src/github.com/docker/editions/tools/branchbot
WORKDIR /go/src/github.com/docker/editions/tools/branchbot
RUN vndr
RUN ["go", "install", "github.com/docker/editions/tools/branchbot"]



FROM alpine

RUN apk add --update --no-cache git openssh
COPY --from=builder /go/bin/branchbot /go/bin/branchbot
RUN chmod +x /go/bin/branchbot
RUN adduser -D branchbot
WORKDIR /home/branchbot
RUN mkdir .ssh
COPY .gitconfig .gitconfig
COPY ssh_config .ssh/config
RUN chmod 400 .ssh/config
RUN chown -R branchbot:branchbot .
USER branchbot
ENV PATH=/go/bin:$PATH
CMD branchbot -token "$(cat /run/secrets/gh-cherrypick-token)" -verbose

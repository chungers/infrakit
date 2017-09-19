FROM golang:1.8.3-alpine


RUN apk update && apk upgrade && \
    apk add --no-cache git 

RUN go get -u github.com/linuxkit/rtf


FROM alpine:3.6

RUN apk update && apk upgrade && apk add --no-cache openssh

COPY --from=0 /go/bin/rtf /usr/local/bin

ARG DOCKER_VERSION
ENV EXPECTED_VERSION ${DOCKER_VERSION}

ARG KEY_PATH
ENV SSH_KEY ${KEY_PATH}

RUN mkdir -p /tests/cases
COPY cases /tests/cases
WORKDIR /tests/

ENTRYPOINT ["rtf"]
CMD ["help"]

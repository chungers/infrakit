FROM alpine:3.5

RUN \
  apk update && apk upgrade && \
  apk add --no-cache \
    curl \
    e2fsprogs \
    groff \
    jq \
    python3 \
    syslinux \
  && true

RUN pip3 install -U awscli
ENV PAGER more
ENTRYPOINT ["aws"]
CMD ["help"]

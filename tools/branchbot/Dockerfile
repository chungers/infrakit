FROM alpine

RUN apk add --update --no-cache git openssh
COPY ./branchbot /go/bin/branchbot
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

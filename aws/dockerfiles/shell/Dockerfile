FROM alpine:3.6
MAINTAINER KenCochrane <Ken.Cochrane@Docker.com>

ARG COMPOSE_VERSION=1.15.0
ENV GLIBC 2.23-r3

RUN apk add --update --no-cache openssh ca-certificates bash sudo curl jq

RUN curl -sSL -o /etc/apk/keys/sgerrand.rsa.pub https://raw.githubusercontent.com/sgerrand/alpine-pkg-glibc/master/sgerrand.rsa.pub && \
    curl -sSL -O https://github.com/sgerrand/alpine-pkg-glibc/releases/download/$GLIBC/glibc-$GLIBC.apk && \
    apk add --no-cache --virtual glibc-compat glibc-$GLIBC.apk && \
    rm glibc-$GLIBC.apk


# Add docker-compose
RUN ln -s /lib/libz.so.1 /usr/glibc-compat/lib/ && \
    ln -s /lib/libc.musl-x86_64.so.1 /usr/glibc-compat/lib

RUN curl -sSL -o /usr/local/bin/docker-compose https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-Linux-x86_64 

COPY etc/ /etc/

COPY files/entry.sh /entry.sh
COPY bin/* /usr/bin/
RUN chmod 755 /entry.sh /usr/bin/swarm-exec /usr/bin/docker-diagnose /usr/local/bin/docker-compose

# Add the docker client to the container (don't rely on host binary)
ARG DOCKER_BIN_URL=http://docker-ci-artifacts.s3.amazonaws.com/ci.qa.aws.dckr.io/jenkins-docker-release-packaging-ce-17.06-12/docker-amd64.tgz

ADD $DOCKER_BIN_URL /docker-amd64.tgz
RUN tar -zxvf /docker-amd64.tgz --strip-components 1 -C /usr/bin docker/docker
RUN rm /docker-amd64.tgz

# add docker to sudoers file
RUN echo -e "\ndocker ALL=(ALL) NOPASSWD: ALL\n" >> /etc/sudoers
RUN echo -e "Port 22\nClientAliveInterval 180\nPasswordAuthentication no\n" >> /etc/ssh/sshd_config

EXPOSE 22

ENTRYPOINT ["/entry.sh"]

CMD ["/usr/sbin/sshd", "-D", "-f", "/etc/ssh/sshd_config"]

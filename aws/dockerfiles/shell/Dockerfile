FROM alpine:3.5

MAINTAINER KenCochrane <Ken.Cochrane@Docker.com>
RUN apk add --update bash openssh ca-certificates sudo curl jq && \
    echo -e "Port 22\nClientAliveInterval 180\nPasswordAuthentication no\n" >> /etc/ssh/sshd_config && \
    rm -rf /var/cache/apk/*

# add docker to sudoers file
RUN echo -e "\ndocker ALL=(ALL) NOPASSWD: ALL\n" >> /etc/sudoers

COPY etc/ /etc/

COPY files/entry.sh /entry.sh
COPY bin/docker-diagnose /usr/bin/docker-diagnose
COPY bin/swarm-exec /usr/bin/swarm-exec
RUN chmod 755 /entry.sh
RUN chmod 755 /usr/bin/swarm-exec
RUN chmod 755 /usr/bin/docker-diagnose

EXPOSE 22

ENTRYPOINT ["/entry.sh"]

CMD ["/usr/sbin/sshd", "-D", "-f", "/etc/ssh/sshd_config"]

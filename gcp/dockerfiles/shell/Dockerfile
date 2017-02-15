FROM alpine:3.5

RUN apk add --update bash openssh ca-certificates sudo curl jq tar python py-setuptools shadow iproute2

CMD ["/entry.sh", "/usr/sbin/sshd", "-D", "-f", "/etc/ssh/sshd_config"]
EXPOSE 22

RUN groupadd google-sudoers -g 50
RUN mkdir -p /src \
  && cd /src \
  && (curl -sSL https://github.com/GoogleCloudPlatform/compute-image-packages/archive/c4793eaac79a1c83c2911e9abb3466e82542a0f8.tar.gz | tar zxv --strip=1) \
  && python setup.py install \
  && cd .. \
  && rm -Rf /src

COPY etc /etc/
COPY entry.sh /
COPY bin/swarm-exec bin/ssh-keygen.sh bin/docker-diagnose /usr/bin/

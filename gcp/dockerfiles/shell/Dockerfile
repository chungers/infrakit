# Switched to debian from alpine, due to bug in usermod: http://bugs.alpinelinux.org/issues/7390
FROM debian:9 

RUN apt-get update && apt-get install -y bash openssh-server ca-certificates sudo curl jq tar python python-setuptools iproute2 && rm -rf /var/lib/apt/lists/*

CMD ["/entry.sh", "/usr/sbin/sshd", "-D", "-f", "/etc/ssh/sshd_config"]
EXPOSE 22

# on debian GID is staff - mod to sudoers
RUN groupmod -n google-sudoers -g 50 staff
RUN mkdir -p /src \
  && cd /src \
  && (curl -sSL https://github.com/GoogleCloudPlatform/compute-image-packages/archive/c4793eaac79a1c83c2911e9abb3466e82542a0f8.tar.gz | tar zxv --strip=1) \
  && python setup.py install \
  && cd .. \
  && rm -Rf /src

COPY etc /etc/
COPY entry.sh /
COPY bin/swarm-exec bin/ssh-keygen.sh bin/docker-diagnose /usr/bin/

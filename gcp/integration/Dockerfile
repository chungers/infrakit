FROM google/cloud-sdk@sha256:fb904276e8a902ccd9564989d9222bdfbe37ffcd7f9989ca7e24b4019a9b4b6b

RUN apt-get update && apt-get install -y \
  curl \
  jq

CMD ["/e2e.sh"]
COPY e2e.sh /
